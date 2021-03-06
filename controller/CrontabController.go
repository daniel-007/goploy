package controller

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"goploy/core"
	"goploy/model"
	"goploy/utils"
	"net/http"
	"strconv"
	"strings"
)

// Crontab struct
type Crontab Controller

// GetList crontab list
func (crontab Crontab) GetList(w http.ResponseWriter, gp *core.Goploy) *core.Response {
	type RespData struct {
		Crontabs   model.Crontabs   `json:"crontabList"`
		Pagination model.Pagination `json:"pagination"`
	}
	pagination, err := model.PaginationFrom(gp.URLQuery)
	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}

	crontabList, pagination, err := model.Crontab{}.GetList(pagination)

	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	return &core.Response{Data: RespData{Crontabs: crontabList, Pagination: pagination}}
}

// GetList crontab list
func (crontab Crontab) GetRemoteServerList(w http.ResponseWriter, gp *core.Goploy) *core.Response {
	type RespData struct {
		Crontabs []string `json:"crontabList"`
	}

	serverID, err := strconv.ParseInt(gp.URLQuery.Get("serverId"), 10, 64)
	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}

	server, err := model.Server{
		ID: serverID,
	}.GetData()

	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}

	session, err := utils.ConnectSSH(server.Owner, "", server.IP, server.Port)

	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}

	var sshOutbuf, sshErrbuf bytes.Buffer
	session.Stdout = &sshOutbuf
	session.Stderr = &sshErrbuf
	sshOutbuf.Reset()
	if err = session.Run("crontab -l"); err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	crontabList := strings.Split(sshOutbuf.String(), "\n")

	var crontabs []string
	for _, crontab := range crontabList {
		if len(crontab) == 0 {
			continue
		}
		// skip comment
		if strings.Index("#", crontab) == 0 {
			continue
		}
		crontabs = append(crontabs, crontab)
	}

	return &core.Response{Data: RespData{crontabs}}
}

// Add one crontab
func (crontab Crontab) Add(w http.ResponseWriter, gp *core.Goploy) *core.Response {
	type ReqData struct {
		Command string `json:"command" validate:"required"`
	}
	type RespData struct {
		ID int64 `json:"id"`
	}
	var reqData ReqData
	if err := verify(gp.Body, &reqData); err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}

	crontabInfo, err := model.Crontab{Command: reqData.Command}.GetDataByCommand()
	if err != nil && err != sql.ErrNoRows {
		return &core.Response{Code: core.Error, Message: err.Error()}
	} else if crontabInfo != (model.Crontab{}) {
		return &core.Response{Code: core.Error, Message: "Command is already exist"}
	}

	id, err := model.Crontab{
		Command:   reqData.Command,
		Creator:   gp.UserInfo.Name,
		CreatorID: gp.UserInfo.ID,
	}.AddRow()

	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}

	}
	return &core.Response{Data: RespData{ID: id}}
}

// Edit one crontab
func (crontab Crontab) Edit(w http.ResponseWriter, gp *core.Goploy) *core.Response {
	type ReqData struct {
		ID      int64  `json:"id" validate:"gt=0"`
		Command string `json:"command" validate:"required"`
	}
	var reqData ReqData
	if err := verify(gp.Body, &reqData); err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}

	crontabInfo, err := model.Crontab{Command: reqData.Command}.GetDataByCommand()
	if err != nil {
		if err != sql.ErrNoRows {
			return &core.Response{Code: core.Error, Message: err.Error()}
		}
	} else if crontabInfo.ID != reqData.ID {
		return &core.Response{Code: core.Error, Message: "Command is already exist"}
	}

	err = model.Crontab{
		ID:       reqData.ID,
		Command:  reqData.Command,
		Editor:   gp.UserInfo.Name,
		EditorID: gp.UserInfo.ID,
	}.EditRow()

	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	return &core.Response{}
}

// import many crontab
func (crontab Crontab) Import(w http.ResponseWriter, gp *core.Goploy) *core.Response {
	type ReqData struct {
		Commands []string `json:"commands" validate:"required"`
	}
	var reqData ReqData
	if err := verify(gp.Body, &reqData); err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}

	commands := make(map[string]string)
	var commandMD5s []string
	for _, command := range reqData.Commands {
		if len(command) == 0 {
			continue
		}
		h := md5.New()
		h.Write([]byte(command))
		commandMD5 := hex.EncodeToString(h.Sum(nil))
		commands[commandMD5] = command
		commandMD5s = append(commandMD5s, commandMD5)
	}

	crontabList, err := model.Crontab{}.GetListInCommandMD5(commandMD5s)
	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}

	for _, crontab := range crontabList {
		if _, ok := commands[crontab.CommandMD5]; ok {
			delete(commands, crontab.CommandMD5)
		}
	}

	var addCommands []string
	for _, command := range commands {
		addCommands = append(addCommands, command)
	}
	if len(addCommands) != 0 {
		err := model.Crontab{Creator: gp.UserInfo.Name, CreatorID: gp.UserInfo.ID}.AddRowsInCommand(addCommands)
		if err != nil {
			return &core.Response{Code: core.Error, Message: err.Error()}
		}
	}

	return &core.Response{}
}

// Remove one Crontab
func (crontab Crontab) Remove(w http.ResponseWriter, gp *core.Goploy) *core.Response {
	type ReqData struct {
		ID int64 `json:"id" validate:"gt=0"`
	}
	var reqData ReqData
	if err := verify(gp.Body, &reqData); err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	err := model.Crontab{
		ID: reqData.ID,
	}.Remove()

	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	return &core.Response{}
}
