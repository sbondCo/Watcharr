import axios from "axios";

const { MODE } = import.meta.env;

axios.interceptors.request.use(
  (config) => {
    if (!config.baseURL) {
      config.baseURL = MODE === "development" ? "http://127.0.0.1:3080" : "/api";
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);
