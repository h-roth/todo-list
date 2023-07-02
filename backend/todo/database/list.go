package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dackroyd/todo-list/backend/todo"
)

type ListRepository struct {
	db *sql.DB
}

func NewListRepository(db *sql.DB) *ListRepository {
	return &ListRepository{db: db}
}

func (r *ListRepository) Items(ctx context.Context, listID string) ([]todo.Item, error) {
	query := `
		-- Name: TODO List Items
		SELECT id,
		       description,
		       due,
		       completed
		  FROM items
		 WHERE list_id = $1
	 `

	cols := func(i *todo.Item) []any {
		return []any{&i.ID, &i.Description, &i.Due, &i.Completed}
	}

	items, err := queryRows(ctx, r.db, cols, query, listID)
	if err != nil {
		return nil, fmt.Errorf("failed to query for list items: %w", err)
	}

	return items, nil
}

func (r *ListRepository) List(ctx context.Context, listID string) (*todo.DueList, error) {
	query := `
		-- Name: TODO List
		SELECT id,
		       description
		  FROM lists
		  WHERE id = $1
	`

	cols := func(l *todo.List) []any {
		return []any{&l.ID, &l.Description}
	}

	list, err := queryRow(ctx, r.db, cols, query, listID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, todo.NotFoundError(fmt.Sprintf("list with id %q does not exist", listID))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query todo list: %w", err)
	}

	due, err := r.dueItems(ctx, listID)
	if err != nil {
		return nil, err
	}

	return &todo.DueList{DueItems: due, List: *list}, nil
}

type nullableItem struct {
	ID          sql.NullString
	Description sql.NullString
	Due         sql.NullTime
	Completed   sql.NullTime
}

func (r *ListRepository) Lists(ctx context.Context) ([]todo.DueList, error) {
	query := `
		SELECT 
			lists.id,
			lists.description,
			items.id,
			items.description,
			items.due,
			items.completed
		FROM lists
		LEFT JOIN (
			SELECT 
				id,
				description,
				due,
				completed,
				list_id
			FROM items
			WHERE due <= now() + INTERVAL '1 day' 
				AND completed IS NULL
		) AS items ON lists.id = items.list_id
		ORDER BY items.due
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query todo lists and due items: %w", err)
	}
	defer rows.Close()

	dueLists := make(map[string]todo.DueList)
	for rows.Next() {
		var list todo.List
		var nullableItem nullableItem
		err := rows.Scan(&list.ID, &list.Description, &nullableItem.ID, &nullableItem.Description, &nullableItem.Due, &nullableItem.Completed)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		dl, ok := dueLists[list.ID]
		if !ok {
			dl = todo.DueList{
				List:     list,
				DueItems: []todo.Item{},
			}
		}

		// Check if item is not NULL before appending
		if nullableItem.ID.Valid {
			item := todo.Item{
				ID:          nullableItem.ID.String,
				Description: nullableItem.Description.String,
				Due:         &nullableItem.Due.Time,
				Completed:   &nullableItem.Completed.Time,
			}
			dl.DueItems = append(dl.DueItems, item)
		}

		dueLists[list.ID] = dl
	}

	// Convert map to slice
	result := make([]todo.DueList, 0, len(dueLists))
	for _, dl := range dueLists {
		result = append(result, dl)
	}

	return result, nil
}

func (r *ListRepository) dueItems(ctx context.Context, listID string) ([]todo.Item, error) {
	query := `
		-- Name: TODO Due List Items
		SELECT id,
		       description,
		       due,
		       completed
		  FROM items
		 WHERE list_id = $1
		   AND due <= now() + INTERVAL '1 day'
		   AND completed IS NULL
		 ORDER BY due
	 `

	itemCols := func(i *todo.Item) []any {
		return []any{&i.ID, &i.Description, &i.Due, &i.Completed}
	}

	items, err := queryRows(ctx, r.db, itemCols, query, listID)
	if err != nil {
		return nil, fmt.Errorf("failed to query due items for todo list %q: %w", listID, err)
	}

	return items, nil
}
