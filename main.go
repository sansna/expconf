package main

import (
	//"strconv"
	"encoding/json"
	"fmt"
	//"time"

	"github.com/gin-gonic/gin"
	//"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/sansna/expconf/api"
	"github.com/sansna/expconf/models"
	"github.com/sansna/expconf/proto"
	"github.com/sansna/expconf/utils/logger"
)

func main() {
	fun := "main"
	err := logger.InitLogger("dev")
	if err != nil {
		fmt.Print(err)
		return
	}
	defer logger.Exit()

	models.InitDB(models.DEFAULT_DB_NAME)

	logger.Debugf("%s: got dbconfs: %v", fun, models.DB_CONF)

	r := gin.Default()

	r.POST("/expconf/add_config", func(c *gin.Context) {
		d, _ := c.GetRawData()
		param := proto.AddConfigParam{}

		json.Unmarshal([]byte(d), &param)
		logger.Debugf("%s: got param: %v", fun, param)

		api.AddConfig(&param)
		c.JSON(200, nil)
	})
	r.POST("/expconf/mod_config", func(c *gin.Context) {
		d, _ := c.GetRawData()
		param := proto.ModConfigParam{}

		json.Unmarshal([]byte(d), &param)
		logger.Debugf("%s: got param: %v", fun, param)

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
		logger.Debugf("%s: got param: %v", fun, param)

		data, _ := api.GetConfig(param)
		c.JSON(200, gin.H{
			"data": data,
		})
	})

	// Put this as defer func to make sure it runs.
	defer func() {
		for _, db := range models.DB_CONF {
			db.Close()
		}
	}()
	r.Run()

	return
}
