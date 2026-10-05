import request from './request'

export function getMappings(params?: { ontology_id?: number; class_id?: number; datasource_id?: number }) {
  return request.get('/mappings', { params }).then(r => r.data)
}

export function getMapping(id: number) {
  return request.get(`/mappings/${id}`).then(r => r.data)
}

export function createMapping(data: any) {
  return request.post('/mappings', data).then(r => r.data)
}

export function updateMapping(id: number, data: any) {
  return request.put(`/mappings/${id}`, data).then(r => r.data)
}

export function deleteMapping(id: number) {
  return request.delete(`/mappings/${id}`).then(r => r.data)
}

export function validateMapping(id: number) {
  return request.post(`/mappings/${id}/validate`).then(r => r.data)
}

export function suggestMappings(data: { ontology_id: number; class_id: number; datasource_id: number; table_name: string }) {
  return request.post('/mappings/suggest', data).then(r => r.data)
}
