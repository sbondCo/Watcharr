import { watchedList } from "@/store";
import axios from "axios";
import { get } from "svelte/store";
import { notify } from "../util/notify";
import type { Tag } from "@/types";

export async function tagWatched(watchedId: number, tag: Tag): Promise<boolean> {
  // If item is already in watched store, run update request instead
  const wList = get(watchedList);
  const wEntry = wList.find((w) => w.id === watchedId);
  const nid = notify({ text: `Tagging`, type: "loading" });
  if (!wEntry) {
    notify({ id: nid, text: "Failed To Tag! Watched entry not found.", type: "error" });
    return false;
  }
  return await axios
    .post(`/watched/${watchedId}/tag/${tag.id}`)
    .then((resp) => {
      console.log("tagWatched: Status:", resp.status);
      if (!wEntry.tags) {
        wEntry.tags = [tag];
      } else {
        wEntry.tags.push(tag);
      }
      watchedList.update(() => wList);
      notify({ id: nid, text: `Tagged!`, type: "success" });
      return true;
    })
    .catch((err) => {
      console.error("tagWatched: Request failed!", err);
      notify({ id: nid, text: "Failed To Tag!", type: "error" });
      return false;
    });
}

export async function untagWatched(watchedId: number, tag: Tag): Promise<boolean> {
  // If item is already in watched store, run update request instead
  const wList = get(watchedList);
  const wEntry = wList.find((w) => w.id === watchedId);
  const nid = notify({ text: `Untagging`, type: "loading" });
  if (!wEntry) {
    notify({ id: nid, text: "Failed To Untag! Watched entry not found.", type: "error" });
    return false;
  }
  return await axios
    .delete(`/watched/${watchedId}/tag/${tag.id}`)
    .then((resp) => {
      console.log("untagWatched: Status:", resp.status);
      if (wEntry.tags) {
        wEntry.tags = wEntry.tags.filter((t) => t.id !== tag.id);
      }
      watchedList.update(() => wList);
      notify({ id: nid, text: `Untagged!`, type: "success" });
      return true;
    })
    .catch((err) => {
      console.error("untagWatched: Request failed!", err);
      notify({ id: nid, text: "Failed To Untag!", type: "error" });
      return false;
    });
}
