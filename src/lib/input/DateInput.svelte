<script lang="ts">
	import type { SvelteDate } from "svelte/reactivity";

	interface Props {
		date: SvelteDate;
	}

	let { date }: Props = $props();

	const pad = (n: number) => String(n).padStart(2, "0");

	function get(d: Date) {
		return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
	}

	function set(v: string) {
		if (!v) return;
		const [y, m, d] = v.split("-").map(Number);
		console.debug("DateInput: set:", "year:", y, "month:", m, "day:", d);
		date.setFullYear(y, m - 1, d);
	}
</script>

<input type="date" bind:value={() => get(date), set} />
