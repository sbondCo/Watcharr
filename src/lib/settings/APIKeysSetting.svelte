<script lang="ts">
	import Spinner from "@/lib/Spinner.svelte";
	import Modal from "@/lib/Modal.svelte";
	import { notify } from "@/lib/util/notify";
	import axios from "axios";
	import dayjs from "dayjs";
	import relativeTime from "dayjs/plugin/relativeTime";
	import { onMount } from "svelte";

	dayjs.extend(relativeTime);
	type APIKey = {
		id: number;
		createdAt: string;
		lastUsed: string;
		revoked: boolean;
	};
	let keys: APIKey[] = [];
	let loading = true;
	let creating = false;
	let newKey: string | null = null;

	async function fetchKeys() {
		loading = true;
		try {
			keys = (await axios.get<APIKey[]>("/api_keys")).data;
		} catch (err) {
			console.error(err);
			notify({ text: "Failed to load keys", type: "error" });
		} finally {
			loading = false;
		}
	}
	async function createKey() {
		creating = true;
		try {
			const r = await axios.post<{ apiKey: string }>("/api_keys");
			newKey = r.data.apiKey;
			notify({ text: "Key created", type: "success" });
			await fetchKeys();
		} catch (err) {
			console.error(err);
			notify({
				text: err?.response?.data?.error ?? "Failed to create key",
				type: "error",
			});
		} finally {
			creating = false;
		}
	}
	async function revokeKey(id: number) {
		if (!confirm("Revoke this key?")) return;
		try {
			await axios.delete(`/api_keys/${id}`);
			notify({ text: "Key revoked", type: "success" });
			await fetchKeys();
		} catch (err) {
			console.error(err);
			notify({
				text: err?.response?.data?.error ?? "Failed to revoke key",
				type: "error",
			});
		}
	}

	onMount(fetchKeys);
</script>

<h4 class="norm" style="margin-top:20px">API Keys</h4>
<p class="sub">
	Generate keys for webhooks. Keep them secret! You can only have 5 active at a
	time per user.
</p>
<button class="primary" disabled={creating} on:click={createKey}
	>Create Key</button
>
{#if loading}
	<Spinner />
{:else if keys.length === 0}
	<p>No keys created yet.</p>
{:else}
	<table>
		<thead
			><tr><th>ID</th><th>Created</th><th>Last Used</th><th></th></tr></thead
		>
		<tbody>
			{#each keys as k}
				<tr class={k.revoked ? "revoked" : ""}>
					<td>{k.id}</td>
					<td>{dayjs(k.createdAt).format("YYYY-MM-DD")}</td>
					<td
						>{k.lastUsed && !k.lastUsed.startsWith("0001")
							? dayjs(k.lastUsed).fromNow()
							: "Never"}</td
					>
					<td
						><button disabled={k.revoked} on:click={() => revokeKey(k.id)}
							>revoke</button
						></td
					>
				</tr>
			{/each}
		</tbody>
	</table>
{/if}

{#if newKey}
	<Modal on:close={() => (newKey = null)}>
		<h3 class="norm">Copy your new API key</h3>
		<pre>{newKey}</pre>
		<button on:click={() => (newKey = null)}>Close</button>
	</Modal>
{/if}

<style lang="scss">
	.sub {
		font-size: 12px;
		margin: 5px 0 10px 0;
	}
	table {
		width: 100%;
		border-collapse: collapse;
		font-size: 12px;
		th,
		td {
			padding: 4px 6px;
			border-bottom: 1px solid #444;
		}
		.revoked {
			opacity: 0.5;
		}
	}
	button.primary {
		margin: 5px 0 10px 0;
	}
</style>
