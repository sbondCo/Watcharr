import { watchedList } from "@/store";
import type {
  MediaType,
  Watched,
  WatchedAddRequest,
  WatchedStatus,
  WatchedUpdateRequest
} from "@/types";
import axios from "axios";
import { get } from "svelte/store";
import { notify } from "./notify";
const { MODE } = import.meta.env;

export const baseURL = MODE === "development" ? "http://127.0.0.1:3080/api" : "/api";

/**
 *
 * @param contentId TMDB ID
 * @param contentType
 * @param status
 * @param rating
 * @returns
 */
export function updateWatched(
  contentId: number,
  contentType: MediaType,
  status?: WatchedStatus,
  rating?: number,
  thoughts?: string
) {
  // If item is already in watched store, run update request instead
  const wList = get(watchedList);
  const wEntry = wList.find(
    (w) => w.content.tmdbId === contentId && w.content.type === contentType
  );
  if (wEntry?.id) {
    if (!status && !rating && !thoughts) return;
    const obj = {} as WatchedUpdateRequest;
    if (status) obj.status = status;
    if (rating) obj.rating = rating;
    if (thoughts) obj.thoughts = thoughts;
    axios
      .put(`/watched/${wEntry.id}`, obj)
      .then((resp) => {
        if (status) wEntry.status = status;
        if (rating) wEntry.rating = rating;
        if (thoughts) wEntry.thoughts = thoughts;
        if (resp?.data?.newActivity) {
          if (wEntry.activity?.length > 0) {
            wEntry.activity.push(resp.data.newActivity);
          } else {
            wEntry.activity = [resp.data.newActivity];
          }
        }
        watchedList.update((w) => w);
        notify({ text: `Saved!`, type: "success" });
      })
      .catch((err) => {
        console.error(err);
        notify({ text: "Failed To Update!", type: "error" });
      });
    return;
  }
  // Add new watched item
  axios
    .post("/watched", {
      contentId,
      contentType,
      rating,
      status
    } as WatchedAddRequest)
    .then((resp) => {
      console.log("Added watched:", resp.data);
      wList.push(resp.data as Watched);
      watchedList.update(() => wList);
    })
    .catch((err) => {
      console.error(err);
      notify({ text: "Failed To Add!", type: "error" });
    });
}

/**
 * Delete an item from watched list.
 * @param id Watched Entry ID
 */
export function removeWatched(id: number) {
  const wList = get(watchedList);
  const wEntry = wList.find((w) => w.id === id);
  if (!wEntry) {
    console.log("Watched entry does not exist!");
    return;
  }
  axios
    .delete(`/watched/${id}`)
    .then((resp) => {
      console.log("Removed watched:", resp.data);
      const newList = wList.filter((w) => w.id !== id);
      watchedList.update(() => newList);
    })
    .catch((err) => {
      console.error(err);
      notify({ text: "Failed To Remove!", type: "error" });
    });
}

/**
 * For use with routes that don't require authentication (eg login/register)
 */
export const noAuthAxios = axios.create({
  baseURL: baseURL
});
