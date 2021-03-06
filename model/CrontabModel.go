package model

import (
	sq "github.com/Masterminds/squirrel"
)

const crontabTable = "`crontab`"

// Crontab mysql table crontab
type Crontab struct {
	ID         int64  `json:"id"`
	Command    string `json:"command"`
	CommandMD5 string `json:"commandMD5"`
	Creator    string `json:"creator"`
	CreatorID  int64  `json:"creatorId"`
	Editor     string `json:"editor"`
	EditorID   int64  `json:"editorId"`
	InsertTime string `json:"insertTime"`
	UpdateTime string `json:"updateTime"`
}

// Crontabs many crontab
type Crontabs []Crontab

// GetList crontab row
func (c Crontab) GetList(pagination Pagination) (Crontabs, Pagination, error) {
	rows, err := sq.
		Select("id, command, creator, creator_id, editor, editor_id, insert_time, update_time").
		From(crontabTable).
		Limit(pagination.Rows).
		Offset((pagination.Page - 1) * pagination.Rows).
		OrderBy("id DESC").
		RunWith(DB).
		Query()
	if err != nil {
		return nil, pagination, err
	}
	crontabs := Crontabs{}
	for rows.Next() {
		var crontab Crontab
		if err := rows.Scan(&crontab.ID, &crontab.Command, &crontab.Creator, &crontab.CreatorID, &crontab.Editor, &crontab.EditorID, &crontab.InsertTime, &crontab.UpdateTime); err != nil {
			return nil, pagination, err
		}
		crontabs = append(crontabs, crontab)
	}
	err = sq.
		Select("COUNT(*) AS count").
		From(crontabTable).
		RunWith(DB).
		QueryRow().
		Scan(&pagination.Total)
	if err != nil {
		return nil, pagination, err
	}
	return crontabs, pagination, nil
}

// GetList crontab row
func (c Crontab) GetListInCommandMD5(commandMD5s []string) (Crontabs, error) {
	rows, err := sq.
		Select("id, command, command_md5").
		From(crontabTable).
		Where(sq.Eq{"command_md5": commandMD5s}).
		RunWith(DB).
		Query()
	if err != nil {
		return nil, err
	}
	crontabs := Crontabs{}
	for rows.Next() {
		var crontab Crontab
		if err := rows.Scan(&crontab.ID, &crontab.Command, &crontab.CommandMD5); err != nil {
			return nil, err
		}
		crontabs = append(crontabs, crontab)
	}
	return crontabs, nil
}

// GetData add crontab information to c Crontab
func (c Crontab) GetData() (Crontab, error) {
	var crontab Crontab
	err := sq.
		Select("id, command, creator, creator_id, editor, editor_id, insert_time, update_time").
		From(crontabTable).
		Where(sq.Eq{"id": c.ID}).
		OrderBy("id DESC").
		RunWith(DB).
		QueryRow().
		Scan(&crontab.ID, &crontab.Command, &crontab.Creator, &crontab.CreatorID, &crontab.Editor, &crontab.EditorID, &crontab.InsertTime, &crontab.UpdateTime)
	if err != nil {
		return crontab, err
	}
	return crontab, nil
}

// GetData add crontab information to s *Crontab
func (c Crontab) GetDataByCommand() (Crontab, error) {
	var crontab Crontab
	err := sq.
		Select("id, creator, creator_id, editor, editor_id, insert_time, update_time").
		From(crontabTable).
		Where(sq.Eq{"command": c.Command}).
		OrderBy("id DESC").
		RunWith(DB).
		QueryRow().
		Scan(&crontab.ID, &crontab.Creator, &crontab.CreatorID, &crontab.Editor, &crontab.EditorID, &crontab.InsertTime, &crontab.UpdateTime)
	if err != nil {
		return crontab, err
	}
	return crontab, nil
}

// AddRow add one row to table crontab
func (c Crontab) AddRow() (int64, error) {
	result, err := sq.
		Insert(crontabTable).
		Columns("command", "command_md5", "creator", "creator_id").
		Values(c.Command, sq.Expr("md5(?)", c.Command), c.Creator, c.CreatorID).
		RunWith(DB).
		Exec()
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return id, err
}

// AddRowsInCommand add many rows to table crontab
func (c Crontab) AddRowsInCommand(commands []string) error {
	builder := sq.
		Insert(crontabTable).
		Columns("command", "command_md5", "creator", "creator_id")
	for _, command := range commands {
		builder = builder.Values(command, sq.Expr("md5(?)", command), c.Creator, c.CreatorID)
	}
	_, err := builder.RunWith(DB).
		Exec()
	if err != nil {
		return err
	}
	return nil
}

// EditRow edit one row to table crontab
func (c Crontab) EditRow() error {
	_, err := sq.
		Update(crontabTable).
		SetMap(sq.Eq{
			"command":     c.Command,
			"command_md5": sq.Expr("md5(?)", c.Command),
			"editor":      c.Editor,
			"editor_id":   c.EditorID,
		}).
		Where(sq.Eq{"id": c.ID}).
		RunWith(DB).
		Exec()
	return err
}

// Remove Crontab
func (c Crontab) Remove() error {
	_, err := sq.
		Update(projectTable).
		SetMap(sq.Eq{
			"state": Disable,
		}).
		Where(sq.Eq{"id": c.ID}).
		RunWith(DB).
		Exec()
	return err
}
