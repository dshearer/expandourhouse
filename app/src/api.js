import axios from 'axios';

function makeAxiosConfig() {
  const headers = {};
  return {headers: headers};
}

function makeUri(uri) {
  return '/api' + uri;
}

function get(uri) {
  return axios.get(makeUri(uri), makeAxiosConfig());
}

function post(uri, data) {
  return axios.post(makeUri(uri), data, makeAxiosConfig());
}

function put(uri, data) {
  return axios.put(makeUri(uri), data, makeAxiosConfig());
}

function patch(uri, data) {
  return axios.patch(makeUri(uri), data, makeAxiosConfig());
}

function del(uri) {
  return axios.delete(makeUri(uri), makeAxiosConfig());
}

export default {get, post, put, patch, del};
