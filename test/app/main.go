package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	app2 "github.com/owenliang/myf-go/app"
	"github.com/owenliang/myf-go/client/mhttp"
	"github.com/owenliang/myf-go/client/mmongo"
	"github.com/owenliang/myf-go/client/mmysql"
	"github.com/owenliang/myf-go/client/mredis"
	"github.com/owenliang/myf-go/mcontext"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
	"net/url"
	"os"
	"time"
)

func test(context *mcontext.MyfContext) {
	if tran := context.Cat.Top(); tran != nil {
		tran.LogEvent("test", "test")
	}
	context.Gin.String(200, "hello world")

	a := 0
	fmt.Println(1 / a)
}

func testMultiHttp(context *mcontext.MyfContext) {
	baiduReq, _ := mhttp.NewGetRequest("https://www.baidu.com", nil, nil)
	iqiyiReq, _ := mhttp.NewPostRequest("https://iqiyi1111.com", nil, nil, nil)

	respList, _ := mhttp.MultiRequest(context, []*mhttp.Request{baiduReq, iqiyiReq})
	for _, resp := range respList {
		fmt.Println(string(resp.Body), resp.Err)
	}
}

func testHttp(context *mcontext.MyfContext) {
	data := url.Values{}
	data.Set("name", "foo")
	data.Add("surname", "bar")
	data.Set("body", `{"title":"Buy cheese and bread for breakfast."}`)
	result, _ := mhttp.Get(context.Context, "http://userapi.myf.com/user/info", &data, &mhttp.HttpConfig{Timeout: 1000})
	context.Gin.String(200, string(result.Body), result.Err)
}

func testRedis(context *mcontext.MyfContext) {
	userReids, _ := mredis.MyfRedis.Instance(context.Context, "user_redis", mredis.MASTER)

	pipe := userReids.Pipeline()
	pipe.Incr("pipeline_1234")
	pipe.Expire("pipeline_1234", time.Hour)

	pipe.Exec()

	myfGoKey := "myf-go-test-key"

	_, err := userReids.Set(myfGoKey, "value-123456", 0).Result()
	if err != nil {
		return
	}

	val2, _ := userReids.Get(myfGoKey).Result()

	context.Gin.String(200, val2)
}

type Account struct {
	SmmyfID int    `gorm:"column:smmyf_id" json:"smmyf_id"`
	UserID  string `gorm:"column:user_id" json:"user_id"`
}

func testMysql(context *mcontext.MyfContext) {
	// 获取数据库
	dbGroup, err := mmysql.MyfDB.Instance("user_db")

	if err != nil {
		context.Gin.String(500, "%v", err)
		return
	}

	var result int64
	rows := []*Account{}
	{
		// 开启事务
		var tx *mmysql.DBConn
		if tx, err = dbGroup.Begin(context); err != nil {
			context.Gin.String(500, "%v", err)
			return
		}
		defer func() {
			if err != nil {
				tx.Rollback(context)
			}
		}()
		//// 执行写入
		//if result, err = tx.Exec(context, "insert into user(`name`) values(?)", "xiaoming"); err != nil {
		//	context.Gin.String(500, "%v", err)
		//	return
		//}
		// 执行查询
		if err = tx.Query(context, &rows, "select * from account1 where user_id = ?", "25855"); err != nil {
			context.Gin.String(500, "%v", err)
			return
		}
		tx.Commit(context)
	}

	context.Gin.JSON(200, gin.H{
		"exec":  result,
		"query": rows,
	})
}

type Notice struct {
	ID         primitive.ObjectID   `json:"_id"  bson:"_id"`
	Message       string   `json:"message" bson:"message"`
	CreationTime     time.Time `json:"-" bson:"creation_time"`

	//格式化数据
	FormatCreationTime string `json:"creation_time"`
}

// 获取一组MongoID
func getMessageMongoMap(ctx context.Context, msgIDList []string) (err error) {
	userMessageMongo, err := mmongo.MyfMongo.Instance("user_message")
	if err != nil {
		return
	}
	//选择数据表
	collection := userMessageMongo.Collection("notice")

	var objectIDList []primitive.ObjectID
	var objectID primitive.ObjectID
	for _, msgID := range msgIDList {
		if len(msgID) <= 0 {
			continue
		}
		objectID, err = primitive.ObjectIDFromHex(msgID)
		if err != nil {
			fmt.Println(err)
			return
		}
		objectIDList = append(objectIDList, objectID)
	}

	filter := bson.D{
		{
			"_id", bson.D{
			{"$in", objectIDList},
		},
		},
	}

	var cur *mongo.Cursor
	cur, err = collection.Find(ctx, filter)
	if err != nil {
		return
	}

	defer cur.Close(ctx)
	for cur.Next(ctx) {
		row := Notice{}
		//解析
		err = cur.Decode(&row)
		if err != nil { return }

		//转换为当前时区
		row.FormatCreationTime = row.CreationTime.In(time.Local).Format("2006-01-02 15:04:05")
		row.CreationTime.Unix()
		data, _ := json.Marshal(row)

		fmt.Printf("%s\n", string(data))
	}

	return
}

func testMongo(context *mcontext.MyfContext) {
	err := getMessageMongoMap(context.Context, []string{"5d85ab05e34411705a2c786c"})
	fmt.Println(err)
}

func main() {
	// 解析命令行
	flag.Parse()

	// 初始化
	app, err := app2.New()
	if err != nil {
		goto END
	}

	// 注册路由
	{
		app.Gin.GET("/test", app.WithMyfContext(test))
		app.Gin.GET("/redis4", app.WithMyfContext(testRedis))
		app.Gin.GET("/mysql2", app.WithMyfContext(testMysql))
		app.Gin.GET("/http", app.WithMyfContext(testHttp))
		app.Gin.GET("/mongo", app.WithMyfContext(testMongo))
		app.Gin.GET("/multi-http", app.WithMyfContext(testMultiHttp))
	}

	// 拉起框架
	if err = app.Run(); err != nil {
		goto END
	}
	return

END:
	fmt.Fprintln(os.Stderr, "框架启动失败,原因：", err)
}
