package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/redis/go-redis/v9"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "github.com/google/uuid"
)

 
type Piece struct {
    ID            string    `bson:"id" json:"id"`
    NamePiece     string    `bson:"namePiece" json:"namePiece"`
    ImgPiece      string    `bson:"imgPiece" json:"imgPiece"`
    Exist         bool      `bson:"Exist" json:"Exist"`
    Date          string    `bson:"Date" json:"Date"`
    NumberOfPiece int       `bson:"numberOfPiece" json:"numberOfPiece"`
    CreatedAt     time.Time `bson:"createdAt" json:"createdAt"`
    UpdatedAt     time.Time `bson:"updatedAt" json:"updatedAt"`
}

func main() {
    ctx := context.Background()

    // ----- Redis Client -----
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
        DB:   0,
    })
    _, err := rdb.Ping(ctx).Result()
    if err != nil {
        log.Fatal("Redis error:", err)
    }

    // ----- MongoDB Client -----
    clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
    mongoClient, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        log.Fatal(err)
    }
    collection := mongoClient.Database("test").Collection("pieces")

    // ----- Gin Router -----
    router := gin.Default()
    router.Use(func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Next()
    })

    // ----- /v1 : Redis cache first -----
    router.GET("/v1", func(c *gin.Context) {
        val, err := rdb.Get(ctx, "allpostredis").Result()
        if err == redis.Nil {
          
            cursor, err := collection.Find(ctx, bson.M{})
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
                return
            }
            var data []Piece
            if err = cursor.All(ctx, &data); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
                return
            }

       
            for i := range data {
                if data[i].ID == "" {
                    data[i].ID = uuid.New().String()
                }
            }

          
            jsonData, _ := json.Marshal(data)
            rdb.Set(ctx, "allpostredis", jsonData, 0)

            fmt.Println("💾 send from database")
            c.JSON(http.StatusOK, data)
        } else if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
        } else {
            // Cache hit
            var data []Piece
            json.Unmarshal([]byte(val), &data)
            fmt.Println("🚀 send from cache")
            c.JSON(http.StatusOK, data)
        }
    })

    // ----- /v2 : always from MongoDB -----
    router.GET("/v2", func(c *gin.Context) {
        cursor, err := collection.Find(ctx, bson.M{})
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
            return
        }
        var data []Piece
        cursor.All(ctx, &data)
        c.JSON(http.StatusOK, data)
    })

    // ----- /clean : flush Redis -----
    router.GET("/clean", func(c *gin.Context) {
        rdb.FlushAll(ctx)
        c.JSON(http.StatusOK, gin.H{"message": "success"})
    })

    // ----- start server -----
    fmt.Println("Server running on port 3000")
    router.Run(":3000")
}
