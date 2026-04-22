package repository

import (
	"context"
	templatedomain "example.com/taskservice/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, task *templatedomain.Task) (*templatedomain.Task, error)
	GetByID(ctx context.Context, id int64) (*templatedomain.Task, error)
	Update(ctx context.Context, task *templatedomain.Task) (*templatedomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]*templatedomain.Task, error)
	CreateTemplate(ctx context.Context, t *templatedomain.Template) (*templatedomain.Template, error)
	GetActiveTemplates(ctx context.Context) ([]*templatedomain.Template, error)
	UpdateTemplateStatus(ctx context.Context, id int64, active bool) error
}
