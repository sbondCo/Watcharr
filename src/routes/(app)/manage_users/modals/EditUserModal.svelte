<script lang="ts">
	import Checkbox from "@/lib/Checkbox.svelte";
	import DropDown from "@/lib/DropDown.svelte";
	import Modal from "@/lib/Modal.svelte";
	import Setting from "@/lib/settings/Setting.svelte";
	import SettingsList from "@/lib/settings/SettingsList.svelte";
	import { req } from "@/lib/util/api";
	import { ReqerError } from "@/lib/util/fetch";
	import { userHasPermission } from "@/lib/util/helpers";
	import { notify } from "@/lib/util/notify";
	import { UserPermission, UserType, type ManagedUser } from "@/types";

	interface UpdateUserRequest {
		permissions?: number;
		type?: UserType;
	}

	interface Props {
		user: ManagedUser;
		onClose: () => void;
	}

	let { user = $bindable(), onClose }: Props = $props();

	let error: string | undefined = $state();
	let formDisabled = false;

	// Things we have changed
	let changedPerms = false;
	let originalUser = structuredClone(user);

	async function save() {
		const changedType = user.type !== originalUser.type;
		// If nothing changed.. error
		if (!changedPerms && !changedType) {
			error = "Nothing has been changed";
			return;
		}
		if (!error) {
			try {
				const toUpdate: UpdateUserRequest = {};
				if (changedPerms) {
					toUpdate["permissions"] = user.permissions;
				}
				if (changedType) {
					toUpdate["type"] = user.type;
				}
				await req.post(`/server/users/${user.id}`, toUpdate);
				notify({
					type: "success",
					text: "Changes saved!",
				});
				onClose();
			} catch (err: any) {
				console.error("Failed to save user!", err);
				error = ReqerError.getMsg(err, "Failed to save");
			}
		}
	}

	function userTogglePermission(perm: UserPermission) {
		user.permissions ^= perm;
		changedPerms = true;
	}
</script>

<Modal
	title={`Edit User`}
	desc={`Configuring ${user.username}`}
	maxWidth="500px"
	{onClose}
>
	{#if error}
		<span class="error">{error}!</span>
	{/if}

	<SettingsList>
		<h3 class="norm">Permissions</h3>

		<Setting
			title="Admin"
			desc="Give user admin, overrides all other permissions."
			row
		>
			<Checkbox
				name="USER_PERM_ADMIN"
				value={userHasPermission(user.permissions, UserPermission.PERM_ADMIN)}
				toggled={() => {
					userTogglePermission(UserPermission.PERM_ADMIN);
				}}
			/>
		</Setting>

		<Setting
			title="Request Content"
			desc="Give user permission to request content."
			row
		>
			<Checkbox
				name="USER_PERM_REQUEST_CONTENT"
				value={userHasPermission(
					user.permissions,
					UserPermission.PERM_REQUEST_CONTENT,
				)}
				toggled={() => {
					userTogglePermission(UserPermission.PERM_REQUEST_CONTENT);
				}}
			/>
		</Setting>

		<Setting
			title="Auto Approve Content Request"
			desc="Auto approve user's content requests."
			row
		>
			<Checkbox
				name="PERM_REQUEST_CONTENT_AUTO_APPROVE"
				value={userHasPermission(
					user.permissions,
					UserPermission.PERM_REQUEST_CONTENT_AUTO_APPROVE,
				)}
				toggled={() => {
					userTogglePermission(
						UserPermission.PERM_REQUEST_CONTENT_AUTO_APPROVE,
					);
				}}
			/>
		</Setting>

		<h3 class="norm">Other</h3>

		<Setting
			title="User Type"
			desc="The type of this user, affects how they login and certain features. Currently only possible to swap between Watcharr/Proxy types."
		>
			{#if !user.type || user.type === UserType.Proxy}
				<DropDown
					placeholder="Unknown"
					bind:active={user.type}
					options={[
						{
							id: 0,
							value: "Watcharr",
						},
						// {
						//   id: UserType.Jellyfin,
						//   value: "Jellyfin",
						// },
						// {
						//   id: UserType.Plex,
						//   value: "Plex",
						// },
						{
							id: UserType.Proxy,
							value: "Proxy",
						},
					]}
					isDropDownItem={true}
				/>
			{:else}
				<p>This option is not supported for this user type yet.</p>
			{/if}
		</Setting>

		<div class="btns">
			<button onclick={() => save()}>Save</button>
		</div>
	</SettingsList>
</Modal>

<style lang="scss">
	.btns {
		display: flex;
		flex-flow: row;
		gap: 10px;

		:first-child {
			margin-left: auto;
		}

		button {
			width: max-content;
			padding-left: 15px;
			padding-right: 15px;
		}
	}

	.error {
		position: sticky;
		top: 0;
		display: flex;
		justify-content: center;
		width: 100%;
		padding: 10px;
		background-color: rgb(221, 48, 48);
		text-transform: capitalize;
		color: white;
		margin-bottom: 15px;
	}
</style>
