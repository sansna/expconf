package main

import "fmt"
import (
	//"strconv"
	"encoding/json"
	//"time"

	"github.com/gin-gonic/gin"
	//"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/sansna/expconf/api"
	"github.com/sansna/expconf/models"
	"github.com/sansna/expconf/proto"
)

func main() {
	models.InitDB(models.DEFAULT_DB_NAME)

	fmt.Println(models.DB_CONF)

	r := gin.Default()

	r.POST("/expconf/add_config", func(c *gin.Context) {
		d, _ := c.GetRawData()
		param := proto.AddConfigParam{}

		json.Unmarshal([]byte(d), &param)
		fmt.Println(param)

		api.AddConfig(&param)
		c.JSON(200, nil)
	})
	r.POST("/expconf/mod_config", func(c *gin.Context) {
		d, _ := c.GetRawData()
		param := proto.ModConfigParam{}

		json.Unmarshal([]byte(d), &param)
		fmt.Println(param, param.OneModifySt)

		api.ModConfig(&param)
		c.JSON(200, nil)
	})
	r.POST("/expconf/get_groups", func(c *gin.Context) {
		d, _ := c.GetRawData()
		param := &proto.GetGroupsParam{}
		json.Unmarshal(d, param)
		data, _ := api.GetGroups(param)
		c.JSON(200, gin.H{
			"data": data,
		})
	})
	r.POST("/expconf/get_config", func(c *gin.Context) {
		d, _ := c.GetRawData()
		param := &proto.GetConfigParam{}
		json.Unmarshal(d, param)
		data, _ := api.GetConfig(param)
		c.JSON(200, gin.H{
			"data": data,
		})
	})

	r.Run()
	for _, db := range models.DB_CONF {
		db.Close()
	}

	return
}
