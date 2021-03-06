import request from '@/utils/request'

/**
 * @return {Promise}
 */
export function getList(pagination) {
  return request({
    url: '/crontab/getList',
    method: 'get',
    params: { ...pagination }
  })
}

/**
 * @return {Promise}
 */
export function getRemoteServerList(serverId) {
  return request({
    url: '/crontab/getRemoteServerList',
    method: 'get',
    params: { serverId },
    timeout: 0
  })
}

export function add(data) {
  return request({
    url: '/crontab/add',
    method: 'post',
    data
  })
}

export function edit(data) {
  return request({
    url: '/crontab/edit',
    method: 'post',
    data
  })
}

export function importCrontab(data) {
  return request({
    url: '/crontab/import',
    method: 'post',
    data
  })
}

export function remove(id) {
  return request({
    url: '/crontab/remove',
    method: 'delete',
    data: { id }
  })
}
