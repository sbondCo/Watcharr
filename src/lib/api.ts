import type { Rating, WatchedStatus, WatchedUpdateRequest } from "@/types";
import axios from "axios";

export function updateWatched(id: number, status?: WatchedStatus, rating?: Rating) {
  if (!status && !rating) return;
  const obj = {} as WatchedUpdateRequest;
  if (status) obj.status = status;
  if (rating) obj.rating = rating;
  return axios.put(`/watched/${id}`, obj);
}
