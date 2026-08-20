<script lang="ts">
	import type { Activity } from "@/types";
	import Collapsible from "../Collapsible.svelte";

	interface Props {
		activity: Activity;
	}

	let { activity }: Props = $props();
</script>

<Collapsible title="View Raw Activity">
	<div>
		<pre>{JSON.stringify(activity, undefined, 2)}</pre>
	</div>
</Collapsible>

<style lang="scss">
	div {
		overflow: auto;
	}
</style>
