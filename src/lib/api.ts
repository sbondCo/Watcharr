import { goto } from "$app/navigation";
import axios from "axios";

export default async function req(ep: string, method: "GET" | "POST", data?: object) {
  const token = localStorage.getItem("token");
  if (!token) {
    goto("/login?again=1");
  }
  return await axios({
    baseURL: "http://127.0.0.1:3080",
    url: ep,
    method,
    headers: {
      Authorization: token
    },
    [method == "GET" ? "params" : "data"]: data
  });
}
