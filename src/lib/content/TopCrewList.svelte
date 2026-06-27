<script lang="ts">
	import type { TMDBContentCreditsCrew } from "@/types";

	interface Props {
		topCrew: TMDBContentCreditsCrew[];
	}

	let { topCrew }: Props = $props();
</script>

<div class="creators">
	{#each topCrew as crew}
		<div>
			<a href={`/person/${crew.id}`}>{crew.name}</a>
			<span>{crew.job}</span>
		</div>
	{/each}
</div>

<style lang="scss">
	.creators {
		display: flex;
		flex-wrap: wrap;
		justify-content: center;
		gap: 35px;
		margin: 10px 60px;

		div {
			display: flex;
			flex-flow: column;
			min-width: 150px;

			a {
				font-weight: bold;
			}
		}
	}
</style>
