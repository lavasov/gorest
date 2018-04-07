package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/lavasov/gorest/model"
)

func (h *Handler) GetTask(c echo.Context) error {
	id := c.Param("id")
	task := &model.Task{}
	if err := h.DB.Where("is_deleted = ?", 0).First(&task, id).Error; err != nil {
		return echo.ErrNotFound
	}

	return c.JSON(http.StatusOK, task)
}

func (h *Handler) DeleteTask(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	task := &model.Task{ID: id}
	h.DB.Model(&task).Update("IsDeleted", true).Delete(&task)

	return c.JSON(http.StatusNoContent, nil)
}

func (h *Handler) UpdateTask(c echo.Context) error {
	task := new(model.Task)
	if err := c.Bind(&task); err != nil {
		return err
	}

	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if task.ID != id {
		return c.JSON(http.StatusBadRequest, nil)
	}

	h.DB.Save(&task)

	return c.JSON(http.StatusCreated, task)
}

func (h *Handler) CreateTask(c echo.Context) error {
	task := &model.Task{}
	if err := c.Bind(&task); err != nil {
		return err
	}

	h.DB.Create(&task)

	return c.JSON(http.StatusCreated, task)
}
