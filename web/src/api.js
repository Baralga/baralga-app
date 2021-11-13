// Api.js
import axios from "axios";
import { navigate } from "svelte-routing";

const axiosAPI = axios.create({
  baseURL: "",
  withCredentials: true,
});

// Auth and Exception interceptor
axiosAPI.interceptors.response.use(
  function (response) {
    return response;
  },
  function (error) {
    if (error.response.status == 401) {
      navigate("/login", { replace: true });
    }
    //TODO: handle errors
    return Promise.reject(error);
  }
);

// implement a method to execute all the request from here.
const apiRequest = (method, url, request) => {

  const headers = {
    "Content-Type": "application/json;charset=UTF-8",
    Accept: "application/json",
  };

  return axiosAPI({
    method: method,
    url: url,
    data: request,
    headers: headers,
  })
    .then((res) => {
      return Promise.resolve(res.data);
    })
    .catch((err) => {
      return Promise.reject(err);
    });
};

// function to execute the http get request
const get = (url, request) => apiRequest("get", url, request);

// function to execute the http delete request
const deleteRequest = (url, request) => apiRequest("delete", url, request);

// function to execute the http post request
const post = (url, request) => apiRequest("post", url, request);

// function to execute the http put request
const put = (url, request) => apiRequest("put", url, request);

// function to execute the http path request
const patch = (url, request) => apiRequest("patch", url, request);

// expose your method to other services or actions
const API = {
  get,
  delete: deleteRequest,
  post,
  put,
  patch,
};
export default API;
