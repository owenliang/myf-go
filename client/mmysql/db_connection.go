package mmysql

import (
	"bytes"
	"context"
	sql2 "database/sql"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	cat2 "github.com/owenliang/myf-go-cat/cat"
	"github.com/owenliang/myf-go/client/cat"
	"reflect"
	"strings"
	"time"
)

// 单个Mysql连接
type DBConn struct {
	gorm *gorm.DB
	tx   bool // 是否是事务

	// 引用配置， 用于cat埋点
	groupConf    *DBGroupConf
	subGroupConf *DBSubGroupConf
	connConf     *DBConnConf
}

func newDBConnection(groupConf *DBGroupConf, subGroupConf *DBSubGroupConf, connConf *DBConnConf) (dbConn *DBConn, err error) {
	initOptions(groupConf, subGroupConf, connConf)

	DSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		groupConf.Username, groupConf.Password, connConf.Host, connConf.Port, groupConf.Database)

	db, err := gorm.Open("mysql", DSN)
	if err != nil {
		return
	}

	db.DB().SetConnMaxLifetime(time.Duration(subGroupConf.IdleTime) * time.Millisecond)
	db.DB().SetMaxOpenConns(subGroupConf.MaxConn)
	db.DB().SetMaxIdleConns(subGroupConf.IdleConn)

	dbConn = &DBConn{
		gorm:         db,
		groupConf:    groupConf,
		subGroupConf: subGroupConf,
		connConf:     connConf,
	}
	return
}

// 初始化默认值
func initOptions(groupConf *DBGroupConf, subGroupConf *DBSubGroupConf, connConf *DBConnConf)  {
	if subGroupConf.MaxConn <= 0 {
		subGroupConf.MaxConn = DEFAULT_MAX_CONN
	}
	if subGroupConf.MaxConn <= 0 {
		subGroupConf.IdleConn = DEFAULT_IDLE_CONN
	}

}

// SQL查询
func (dbConn *DBConn) Query(context context.Context, result interface{}, sql string, values ...interface{}) (err error) {
	var (
		type1, type2, type3 reflect.Type
	)

	// CAT埋点
	{
		if myfCat, ok := context.Value("cat").(*cat.MyfCat); ok {
			myfCat.Append(cat2.NewTransaction("SQL", "SELECT"))
			myfCat.Top().LogEvent("SQL.Method", "SELECT")
			myfCat.Top().LogEvent("SQL.Database", dbConn.JDBC())
			myfCat.Top().LogEvent("SQL.name", sql)
			defer func() {
				if err != nil {
					myfCat.Top().SetStatus(err.Error())
				}
				myfCat.Pop()
			}()
		}
	}

	if type1 = reflect.TypeOf(result); type1.Kind() != reflect.Ptr { // type1是*[]*Element
		return ERR_QUERY_RESULT_INVALID
	}
	if type2 = type1.Elem(); type2.Kind() != reflect.Slice { // type2是[]*Element
		return ERR_QUERY_RESULT_INVALID
	}
	if type3 = type2.Elem(); type3.Kind() != reflect.Ptr { // type3是*Element
		return ERR_QUERY_RESULT_INVALID
	}

	// 发起SQL查询
	var rows *sql2.Rows
	if rows, err = dbConn.gorm.Raw(sql, values...).Rows(); err != nil {
		return
	}
	for rows.Next() {
		elem := reflect.New(type3.Elem())                                   // 创建*Element
		if err = dbConn.gorm.ScanRows(rows, elem.Interface()); err != nil { // 填充*Element
			return
		}
		newSlice := reflect.Append(reflect.ValueOf(result).Elem(), elem) // 将*Element追加到*result
		reflect.ValueOf(result).Elem().Set(newSlice)                     // 将新slice赋值给*result
	}
	return
}

// SQL写入
func (dbConn *DBConn) Exec(context context.Context, sql string, values ...interface{}) (result int64, err error) {
	var sqlResult sql2.Result
	sqlType := dbConn.sqlType(sql)

	// CAT埋点
	{
		if myfCat, ok := context.Value("cat").(*cat.MyfCat); ok {
			myfCat.Append(cat2.NewTransaction("SQL", sqlType))
			myfCat.Top().LogEvent("SQL.Method", sqlType)
			myfCat.Top().LogEvent("SQL.Database", dbConn.JDBC())
			myfCat.Top().LogEvent("SQL.name", sql)
			defer func() {
				if err != nil {
					myfCat.Top().SetStatus(err.Error())
				}
				myfCat.Pop()
			}()
		}
	}

	// 执行SQL
	if sqlResult, err = dbConn.gorm.CommonDB().Exec(sql, values...); err != nil {
		return
	}

	// 判断SQL类型取不同结果
	if sqlType == "INSERT" {
		result, err = sqlResult.LastInsertId()
	} else {
		result, err = sqlResult.RowsAffected()
	}
	return
}

// 开启事务
func (dbConn *DBConn) Begin(context context.Context) (txConn *DBConn, err error) {
	if dbConn.tx {
		return nil, ERR_RECURSION_TX
	}
	clone := *dbConn
	clone.gorm = dbConn.gorm.BeginTx(context, nil)
	clone.tx = true
	txConn = &clone
	return
}

// 提交事务
func (dbConn *DBConn) Commit(context context.Context) (err error) {
	if !dbConn.tx {
		return ERR_INVALID_TX
	}
	dbConn.gorm.Commit()
	return
}

// 回滚事务
func (dbConn *DBConn) Rollback(context context.Context) (err error) {
	if !dbConn.tx {
		return ERR_INVALID_TX
	}
	dbConn.gorm.Rollback()
	return
}

// 判断SQL类型
func (dbConn *DBConn) sqlType(sql string) string {
	sql = strings.TrimLeft(sql, " \t\r\n")

	buf := bytes.Buffer{}
	for i := 0; i < len(sql); i++ {
		if sql[i] != ' ' && sql[i] != '\t' && sql[i] != '\r' && sql[i] != '\n' {
			buf.WriteByte(sql[i])
		} else {
			break
		}
	}
	return strings.ToUpper(buf.String())
}

// 生成JDBC风格的DSN
func (dbConn *DBConn) JDBC() (dsn string) {
	dsn = fmt.Sprintf("jdbc:mysql://%s:%d/%s?useUnicode=true&characterEncoding=utf8mb4&autoReconnect=true",
		dbConn.connConf.Host, dbConn.connConf.Port, dbConn.groupConf.Database)
	return
}
