<script lang="ts">
	import type { SvelteDate } from "svelte/reactivity";

	interface Props {
		date: SvelteDate;
	}

	let { date }: Props = $props();

	const pad = (n: number) => String(n).padStart(2, "0");

	function get(d: Date) {
		return `${pad(d.getHours())}:${pad(d.getMinutes())}`;
	}

	function set(v: string) {
		if (!v) return;
		const [h, m] = v.split(":").map(Number);
		console.debug("DateInput: set:", "hour:", h, "minute:", m);
		date.setHours(h, m);
	}
</script>

<input type="time" bind:value={() => get(date), set} />
