import { t } from "@/lib/i18n/index.svelte";
import { req } from "../util/api";
import { notify } from "../util/notify";
import type { Tag } from "@/types";

export async function tagWatched(
	watchedId: number,
	tag: Tag,
): Promise<boolean> {
	const nid = notify({ text: t("tags.tagging"), type: "loading" });
	if (typeof watchedId !== "number") {
		notify({
			id: nid,
			text: t("tags.failedToTagNotFound"),
			type: "error",
		});
		return false;
	}
	return await req
		.post(`/watched/${watchedId}/tag/${tag.id}`)
		.then((resp) => {
			console.log("tagWatched: Success.", resp);
			notify({ id: nid, text: t("tags.tagged"), type: "success" });
			return true;
		})
		.catch((err) => {
			console.error("tagWatched: Request failed!", err);
			notify({ id: nid, text: t("tags.failedToTag"), type: "error" });
			return false;
		});
}

export async function untagWatched(
	watchedId: number,
	tag: Tag,
): Promise<boolean> {
	const nid = notify({ text: t("tags.untagging"), type: "loading" });
	if (typeof watchedId !== "number") {
		notify({
			id: nid,
			text: t("tags.failedToUntagNotFound"),
			type: "error",
		});
		return false;
	}
	return await req
		.delete(`/watched/${watchedId}/tag/${tag.id}`)
		.then((resp) => {
			console.log("untagWatched: Success.", resp);
			notify({ id: nid, text: t("tags.untagged"), type: "success" });
			return true;
		})
		.catch((err) => {
			console.error("untagWatched: Request failed!", err);
			notify({ id: nid, text: t("tags.failedToUntag"), type: "error" });
			return false;
		});
}
