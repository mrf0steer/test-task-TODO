package handlers

import (
	"context"
	"strconv"
	m "test-task-TODO/internal/models"
	psql "test-task-TODO/storage"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	db *psql.DB
}

func New(db *psql.DB) *Handler {
	return &Handler{db: db}
}

// GET
func (h *Handler) GetTasks(c *fiber.Ctx) error {
	rows, err := h.db.DB.Query(context.Background(), `SELECT id, title, description, status, created_at, updated_at FROM tasks`)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"Status":  "Error",
			"Message": "Error fetching data",
		})
	}
	defer rows.Close()

	var tasks []m.Task
	for rows.Next() {
		var t m.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt); err == nil {
			tasks = append(tasks, t)
		}
	}

	return c.Status(200).JSON(tasks)
}

// POST
func (h *Handler) PostTask(c *fiber.Ctx) error {
	var t m.Task
	if err := c.BodyParser(&t); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"Status":  "Error",
			"Message": "Bad request"})
	}

	err := h.db.DB.QueryRow(
		context.Background(), `INSERT INTO tasks (title, description, status) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`,
		t.Title, t.Description, t.Status).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"Status":  "Error",
			"Message": "Failed to create task",
		})
	}

	if t.Status != "new" && t.Status != "in_progress" && t.Status != "done" {
		return c.Status(400).JSON(fiber.Map{
			"Status":  "Error",
			"Message": "Invalid status",
		})
	}
	return c.Status(200).JSON(fiber.Map{
		"Status":  "Success",
		"message": "Task add successfully",
	})
}

// PUT
func (h *Handler) PutTask(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"Status":  "Error",
			"Message": "Bad ID"})
	}

	var t m.Task
	if err := c.BodyParser(&t); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"Status":  "Error",
			"Message": "Bad request: Invalid data format"})
	}

	t.UpdatedAt = time.Now()

	_, err = h.db.DB.Exec(
		context.Background(),
		`UPDATE tasks SET title=$1, description=$2, status=$3, updated_at=$4 WHERE id=$5`,
		t.Title, t.Description, t.Status, t.UpdatedAt, id,
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"Status":  "Error",
			"Message": "Update failed",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"Status":  "Success",
		"message": "Task updated successfully",
	})
}

// DELETE
func (h *Handler) DeleteTask(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"Status":  "Error",
			"Message": "Invalid ID",
		})
	}

	result, err := h.db.DB.Exec(context.Background(), `DELETE FROM tasks WHERE id=$1`, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"Status":  "Error",
			"Message": "Could not delete",
		})
	}
	if result.RowsAffected() == 0 {
		return c.Status(404).JSON(fiber.Map{
			"Status":  "Error",
			"Message": "Task not found",
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"Status":  "Success",
		"Message": "Task was deleted",
	})
}
