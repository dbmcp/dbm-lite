import { defineStore } from 'pinia'

interface DsState {
  currentDatasourceId: string
  currentDatabase: string
  currentDatasourceName: string
}

export const useDatasourceStore = defineStore('datasource', {
  state: (): DsState => ({
    currentDatasourceId: localStorage.getItem('dbm_ds') || '',
    currentDatabase: localStorage.getItem('dbm_db') || '',
    currentDatasourceName: ''
  }),
  actions: {
    setDatasource(id: string, name: string) {
      this.currentDatasourceId = id
      this.currentDatasourceName = name
      this.currentDatabase = ''
      localStorage.setItem('dbm_ds', id)
      localStorage.removeItem('dbm_db')
    },
    setDatabase(db: string) {
      this.currentDatabase = db
      localStorage.setItem('dbm_db', db)
    }
  }
})
