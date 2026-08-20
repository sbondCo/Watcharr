<script lang="ts">
	import Icon from "../Icon.svelte";
	import type { Icon as IconT } from "@/types";
	import tooltip from "../actions/tooltip";

	interface Props {
		icon: IconT;
		tooltipText: string;
	}

	let { icon, tooltipText }: Props = $props();
</script>

<i
	use:tooltip={{
		text: tooltipText,
		pos: "top",
	}}
	style="width: 20px; height: 20px;"
>
	<Icon i={icon} wh={20} />
</i>
