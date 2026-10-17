<script lang="ts">
	import Checkbox from "@/lib/Checkbox.svelte";
	import Error from "@/lib/Error.svelte";
	import Setting from "@/lib/settings/Setting.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import { req } from "@/lib/util/api";
	import { toRelativeDate } from "@/lib/util/helpers";
	import { notify } from "@/lib/util/notify";
	import type { AllUserServicesResponse, UserServiceSettings } from "@/types";
	import { onMount } from "svelte";

	let error: unknown | undefined = $state();
	let data: AllUserServicesResponse | undefined = $state();

	let autoSyncDisabled = $state(false);

	async function load() {
		error = undefined;
		data = undefined;
		try {
			data = await req.get<AllUserServicesResponse>(`/user/services`);
		} catch (err) {
			error = err;
		}
	}

	/**
	 * Update a setting on UserService.
	 */
	async function updateSetting(
		id: number,
		body: UserServiceSettings,
	): Promise<UserServiceSettings> {
		try {
			const resp = await req.patch<UserServiceSettings>(
				`/user/services/${id}/settings`,
				body,
			);
			return resp;
		} catch (err) {
			console.error("updating autoSync failed!", err);
			notify({ type: "error", text: "Request failed!" });
			return {};
		}
	}

	onMount(() => {
		load();
	});
</script>

{#if error}
	<Error {error} pretty="Failed getting services!" />
{:else if !data}
	<Spinner />
{:else if data.length > 0}
	{#each data as us (us.id)}
		<div class="service-ctr">
			<div class="title">
				<h3 class="norm">{us.name}</h3>
				<span>
					{us.createdAt ? toRelativeDate(new Date(us.createdAt)) : ""}
				</span>
			</div>

			<Setting title="Auto Sync" desc="Allow automatic sync?" row>
				<Checkbox
					name="autoSync"
					disabled={autoSyncDisabled}
					value={us.autoSync}
					toggled={async (on) => {
						autoSyncDisabled = true;
						const s = await updateSetting(us.id, { autoSync: on });
						if (typeof s.autoSync !== "undefined") {
							us.autoSync = s.autoSync;
						}
						autoSyncDisabled = false;
					}}
				/>
			</Setting>
		</div>
	{/each}
{:else}
	None
{/if}

<style lang="scss">
	div.service-ctr {
		display: flex;
		flex-flow: column;
		gap: 10px;
		border: 2px dashed $text-color;
		border-radius: 10px;
		padding: 15px;

		& > .title {
			display: flex;
			flex-flow: row;
			align-items: center;
			gap: 8px;
			text-transform: capitalize;

			span {
				font-size: 15px;
			}
		}
	}
</style>
