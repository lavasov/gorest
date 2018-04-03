package handler

import (
	"net/http"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/lavasov/gorest/model"
)

func GetTask(db *gorm.DB, c echo.Context) error {
	id := c.Param("id")
	task := &model.Task{}
	if err := db.Where("is_deleted = ?", 0).First(&task, id).Error; err != nil {
		return echo.ErrNotFound
	}

	return c.JSON(http.StatusOK, task)
}

func DeleteTask(db *gorm.DB, c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	task := &model.Task{ID: id}
	db.Model(&task).Update("IsDeleted", true).Delete(&task)

	return c.JSON(http.StatusNoContent, nil)
}

func UpdateTask(db *gorm.DB, c echo.Context) error {
	task := new(model.Task)
	if err := c.Bind(&task); err != nil {
		return err
	}

	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if task.ID != id {
		return c.JSON(http.StatusBadRequest, nil)
	}

	db.Save(&task)

	return c.JSON(http.StatusCreated, task)
}

func CreateTask(db *gorm.DB, c echo.Context) error {
	task := &model.Task{}
	if err := c.Bind(&task); err != nil {
		return err
	}

	db.Create(&task)

	return c.JSON(http.StatusCreated, task)
}
