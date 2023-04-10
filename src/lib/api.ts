import { goto } from "$app/navigation";
import axios from "axios";
const { MODE } = import.meta.env;

export default async function req(ep: string, method: "GET" | "POST" | "PUT", data?: object) {
  const token = localStorage.getItem("token");
  if (!token) {
    goto("/login?again=1");
  }
  return await axios({
    baseURL: MODE === "development" ? "http://127.0.0.1:3080" : "/api",
    url: ep,
    method,
    headers: {
      Authorization: token
    },
    [method == "GET" ? "params" : "data"]: data
  });
}
