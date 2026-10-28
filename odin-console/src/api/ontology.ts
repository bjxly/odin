import request from './request'

// 本体 CRUD
export function getOntologies(params?: { current?: number; size?: number }) {
  return request.get('/ontologies', { params }).then(r => r.data)
}

export function getOntology(id: number) {
  return request.get(`/ontologies/${id}`).then(r => r.data)
}

export function createOntology(data: any) {
  return request.post('/ontologies', data).then(r => r.data)
}

export function updateOntology(id: number, data: any) {
  return request.put(`/ontologies/${id}`, data).then(r => r.data)
}

export function deleteOntology(id: number) {
  return request.delete(`/ontologies/${id}`).then(r => r.data)
}

// Classes 子资源
export function getClasses(ontologyId: number) {
  return request.get(`/ontologies/${ontologyId}/classes`).then(r => r.data)
}

export function createClass(ontologyId: number, data: any) {
  return request.post(`/ontologies/${ontologyId}/classes`, data).then(r => r.data)
}

export function updateClass(ontologyId: number, classId: number, data: any) {
  return request.put(`/ontologies/${ontologyId}/classes/${classId}`, data).then(r => r.data)
}

export function deleteClass(ontologyId: number, classId: number) {
  return request.delete(`/ontologies/${ontologyId}/classes/${classId}`).then(r => r.data)
}

// Properties 子资源
export function getProperties(ontologyId: number, classId: number) {
  return request.get(`/ontologies/${ontologyId}/classes/${classId}/properties`).then(r => r.data)
}

export function createProperty(ontologyId: number, classId: number, data: any) {
  return request.post(`/ontologies/${ontologyId}/classes/${classId}/properties`, data).then(r => r.data)
}

export function updateProperty(ontologyId: number, classId: number, propId: number, data: any) {
  return request.put(`/ontologies/${ontologyId}/classes/${classId}/properties/${propId}`, data).then(r => r.data)
}

export function deleteProperty(ontologyId: number, classId: number, propId: number) {
  return request.delete(`/ontologies/${ontologyId}/classes/${classId}/properties/${propId}`).then(r => r.data)
}

// Relations 子资源
export function getRelations(ontologyId: number) {
  return request.get(`/ontologies/${ontologyId}/relations`).then(r => r.data)
}

export function createRelation(ontologyId: number, data: any) {
  return request.post(`/ontologies/${ontologyId}/relations`, data).then(r => r.data)
}

export function updateRelation(ontologyId: number, relId: number, data: any) {
  return request.put(`/ontologies/${ontologyId}/relations/${relId}`, data).then(r => r.data)
}

export function deleteRelation(ontologyId: number, relId: number) {
  return request.delete(`/ontologies/${ontologyId}/relations/${relId}`).then(r => r.data)
}

// Rules 子资源
export function getRules(ontologyId: number) {
  return request.get(`/ontologies/${ontologyId}/rules`).then(r => r.data)
}

export function createRule(ontologyId: number, data: any) {
  return request.post(`/ontologies/${ontologyId}/rules`, data).then(r => r.data)
}

export function updateRule(ontologyId: number, ruleId: number, data: any) {
  return request.put(`/ontologies/${ontologyId}/rules/${ruleId}`, data).then(r => r.data)
}

export function deleteRule(ontologyId: number, ruleId: number) {
  return request.delete(`/ontologies/${ontologyId}/rules/${ruleId}`).then(r => r.data)
}
