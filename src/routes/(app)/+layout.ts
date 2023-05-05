export const prerender = false;
export const ssr = false;
export const csr = true;

import { goto } from "$app/navigation";
import axios from "axios";
import { baseURL } from "@/lib/api";

axios.interceptors.request.use(
  (config) => {
    if (!config.baseURL) {
      config.baseURL = baseURL;

      // Only want to set auth header if requesting to our backend.
      const token = localStorage.getItem("token");
      // Don't require token check if going to auth route (login/register)
      if (!token && !config.url?.includes("/auth")) {
        console.error("No token, going to login. Endpoint:", config.url);
        goto("/login?again=1");
        throw new axios.Cancel("No auth token found");
      }
      config.headers.set("Authorization", token);
    }

    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

axios.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    if (error.response?.status === 401) {
      console.error("Recieved 401 response, going to login.");
      localStorage.removeItem("token");
      goto("/login?again=1");
    }
    return Promise.reject(error);
  }
);
