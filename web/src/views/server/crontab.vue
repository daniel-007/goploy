<template>
  <el-row class="app-container">
    <el-row class="app-bar" type="flex" justify="end">
      <el-button type="primary" icon="el-icon-download" @click="handleImport">导入</el-button>
      <el-button type="primary" icon="el-icon-plus" @click="handleAdd">添加</el-button>
    </el-row>
    <el-table
      border
      stripe
      highlight-current-row
      :data="tableData"
      style="width: 100%"
    >
      <el-table-column prop="id" label="ID" min-width="50" />
      <el-table-column prop="command" label="命令" min-width="140" show-overflow-tooltip />
      <el-table-column prop="description" label="描述" min-width="240" show-overflow-tooltip />
      <el-table-column prop="creator" label="创建人" min-width="50" />
      <el-table-column prop="editor" label="修改人" min-width="50" />
      <el-table-column prop="insertTime" label="创建时间" width="135" />
      <el-table-column prop="updateTime" label="更新时间" width="135" />
      <el-table-column prop="operation" label="操作" width="150" fixed="right">
        <template slot-scope="scope">
          <el-button size="small" type="primary" @click="handleEdit(scope.row)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleRemove(scope.row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-row type="flex" justify="end" style="margin-top: 10px;">
      <el-pagination
        hide-on-single-page
        :total="pagination.total"
        :page-size="pagination.rows"
        background
        layout="prev, pager, next"
        @current-change="handlePageChange"
      />
    </el-row>
    <el-dialog title="crontab设置" :visible.sync="dialogVisible">
      <el-form ref="form" v-loading="formProps.loading" :rules="formRules" :model="formData" label-width="80px">
        <el-form-item label="时间" prop="date">
          <el-input v-model.trim="formData.date" autocomplete="off" placeholder="* * * * ?" @change="onDateChange" />
          <span>{{ formProps.dateCN }}</span>
        </el-form-item>
        <el-form-item label="脚本" prop="script">
          <el-input v-model.trim="formData.script" autocomplete="off" />
        </el-form-item>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button @click="dialogVisible = false">取 消</el-button>
        <el-button :disabled="formProps.disabled" type="primary" @click="submit">确 定</el-button>
      </div>
    </el-dialog>
    <el-dialog title="crontab导入" :visible.sync="importVisible">
      <el-row>
        <el-row type="flex">
          <el-select v-model="importProps.serverId" placeholder="选择服务器" style="width:100%;margin-right:5px;">
            <el-option
              v-for="(item, index) in serverOption"
              :key="index"
              :label="item.label"
              :value="item.id"
            />
          </el-select>
          <el-button :disabled="importProps.disabled" :icon="importProps.disabled ? 'el-icon-loading' : 'el-icon-search'" type="primary" @click="getRemoteServerList">读取</el-button>
        </el-row>
        <el-table
          border
          stripe
          highlight-current-row
          empty-text="请先选择服务器读取"
          :data="serverTableData"
          style="width: 100%; margin-top: 10px;"
          @selection-change="handleCrontabSelectionChange"
        >
          <el-table-column
            type="selection"
            width="40"
          />
          <el-table-column prop="command" label="命令" min-width="140" show-overflow-tooltip />
          <el-table-column prop="description" label="描述" min-width="240" show-overflow-tooltip />
        </el-table>
      </el-row>
      <div slot="footer" class="dialog-footer">
        <el-button @click="dialogVisible = false">取 消</el-button>
        <el-button :disabled="importProps.disabled" type="primary" @click="importCrontab">确 定</el-button>
      </div>
    </el-dialog>
  </el-row>
</template>
<script>
import cronstrue from 'cronstrue/i18n'
import { getList, getRemoteServerList, add, edit, remove, importCrontab } from '@/api/crontab'
import { getOption as getServerOption } from '@/api/server'

export default {
  data() {
    const validateDate = (rule, value, callback) => {
      try {
        cronstrue.toString(value, { locale: 'zh_CN' })
        callback()
      } catch (error) {
        callback(error)
      }
    }
    return {
      dialogVisible: false,
      importVisible: false,
      selectedItems: [],
      tableData: [],
      serverOption: [],
      pagination: {
        page: 1,
        rows: 16,
        total: 0
      },
      formProps: {
        loading: false,
        disabled: false,
        dateCN: ''
      },
      formData: {
        id: 0,
        command: '',
        date: '',
        script: ''
      },
      formRules: {
        date: [
          { required: true, validator: validateDate, trigger: 'blur' }
        ],
        script: [
          { required: true, message: '请输入脚本', trigger: 'blur' }
        ]
      },
      importProps: {
        serverId: '',
        disabled: false,
        loading: false
      },
      serverTableData: []
    }
  },
  created() {
    this.getList()
    this.getServerOption()
  },

  methods: {
    getList() {
      getList(this.pagination).then((response) => {
        const crontabList = response.data.crontabList || []
        this.tableData = crontabList.map(element => {
          const commandSplit = element.command.split(' ')
          element.date = commandSplit.slice(0, 5).join(' ')
          element.dateCN = cronstrue.toString(element.date, { locale: 'zh_CN' })
          element.script = commandSplit.slice(5).join(' ')
          element.description = element.dateCN + ', 运行: ' + element.script
          return element
        })
        this.pagination = response.data.pagination
      })
    },

    getServerOption() {
      getServerOption().then((response) => {
        this.serverOption = response.data.serverList || []
        this.serverOption.map(element => {
          element.label = element.name + (element.description.length > 0 ? '(' + element.description + ')' : '')
          return element
        })
      })
    },

    getRemoteServerList() {
      if (this.importProps.serverId <= 0) {
        this.$message.warning('请先选择服务器')
        return
      }
      this.importProps.disabled = true
      getRemoteServerList(this.importProps.serverId).then(response => {
        this.serverTableData = response.data.crontabList.map(command => {
          const element = {}
          const commandSplit = command.split(' ')
          element.command = command
          element.date = commandSplit.slice(0, 5).join(' ')
          element.dateCN = cronstrue.toString(element.date, { locale: 'zh_CN' })
          element.script = commandSplit.slice(5).join(' ')
          element.description = element.dateCN + ', 运行: ' + element.script
          return element
        })
      }).finally(() => { this.importProps.disabled = false })
    },

    handleAdd() {
      this.formData.id = 0
      this.dialogVisible = true
    },

    handleImport() {
      this.importVisible = true
    },

    handleEdit(data) {
      this.formData.id = data.id
      this.formData.date = data.date
      this.formData.script = data.script
      this.formProps.dateCN = data.dateCN
      this.dialogVisible = true
    },

    handleRemove(data) {
      this.$confirm('此操作将删除该定时任务, 是否继续?', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        remove(data.id).then((response) => {
          this.$message({
            message: '删除成功',
            type: 'success',
            duration: 5 * 1000
          })
          this.getList()
        })
      }).catch(() => {
        this.$message({
          type: 'info',
          message: '已取消删除'
        })
      })
    },

    handleCrontabSelectionChange(items) {
      this.selectedItems = items
    },

    importCrontab() {
      if (this.selectedItems.length === 0) {
        this.$message.warning('请先选择需要导入的条目')
        return
      }
      importCrontab({ commands: this.selectedItems.map(element => element.command) }).then(response => {
        this.$message.success('导入成功')
      }).finally(() => { this.importVisible = false })
    },

    onDateChange() {
      this.formProps.dateCN = cronstrue.toString(this.formData.date, { locale: 'zh_CN' })
    },

    handlePageChange(val) {
      this.pagination.page = val
      this.getProjectList()
    },

    submit() {
      this.$refs.form.validate((valid) => {
        if (valid) {
          this.formData.command = this.formData.date + ' ' + this.formData.script
          if (this.formData.id === 0) {
            this.add()
          } else {
            this.edit()
          }
        } else {
          return false
        }
      })
    },

    add() {
      this.formProps.disabled = true
      add(this.formData).then((response) => {
        this.getList()
        this.$message({
          message: '添加成功',
          type: 'success',
          duration: 5 * 1000
        })
      }).finally(() => {
        this.formProps.disabled = this.dialogVisible = false
      })
    },

    edit() {
      this.formProps.disabled = true
      edit(this.formData).then((response) => {
        this.getList()
        this.$message({
          message: '编辑成功',
          type: 'success',
          duration: 5 * 1000
        })
      }).finally(() => {
        this.formProps.disabled = this.dialogVisible = false
      })
    }
  }
}
</script>
<style lang="scss" scoped>
@import "@/styles/mixin.scss";
.template-dialog {
  padding-right: 10px;
  height: 400px;
  overflow-y: auto;
  @include scrollBar();
}
</style>
