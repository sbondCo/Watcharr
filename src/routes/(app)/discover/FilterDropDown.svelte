<script lang="ts">
	import { t } from "@/lib/i18n/index.svelte";
	import tooltip from "@/lib/actions/tooltip";
	import DropDown from "@/lib/DropDown.svelte";
	import {
		DiscoverFilter,
		SearchType,
		type DiscoverFilterOption,
		type DropDownItem,
	} from "@/types";

	interface Props {
		active: DiscoverFilter | undefined;
		discoverType: SearchType | undefined;
		onChange: () => void;
	}

	let { active = $bindable(), discoverType, onChange }: Props = $props();

	const dropDownOptions: { [x in DiscoverFilterOption]: DropDownItem } =
		$derived({
			trending: {
				id: DiscoverFilter.trending,
				value: t("discover.filter.trending"),
			},
			popular: {
				id: DiscoverFilter.popular,
				value: t("discover.filter.popular"),
			},
			upcoming: {
				id: DiscoverFilter.upcoming,
				value: t("discover.filter.upcoming"),
			},
			intheatres: {
				id: DiscoverFilter.inTheatres,
				value: t("discover.filter.inTheatres"),
			},
			streaming: {
				id: DiscoverFilter.streaming,
				value: t("discover.filter.streaming"),
			},
			// advanced: {
			// 	id: "advanced",
			// 	value: "Advanced Search",
			// },
		});

	let options = $derived.by(() => {
		let o: DropDownItem[] = [dropDownOptions.trending];
		switch (discoverType) {
			case SearchType.movie:
				o.push(
					dropDownOptions.popular,
					dropDownOptions.upcoming,
					dropDownOptions.intheatres,
				);
				break;
			case SearchType.show:
				o.push(dropDownOptions.popular, dropDownOptions.upcoming);
				break;
			case SearchType.person:
				o.push(dropDownOptions.popular);
				break;
			case SearchType.game:
				o.push(dropDownOptions.upcoming);
				break;
		}
		// o.push(dropDownOptions.advanced);
		return o;
	});
	let onMultiDiscover = $derived(discoverType === SearchType.multi);
</script>

<div
	use:tooltip={{
		text: t("discover.mustSelectType"),
		pos: "left",
		condition: onMultiDiscover,
	}}
>
	<DropDown
		placeholder={t("discover.filter.trending")}
		{options}
		isDropDownItem={true}
		bind:active
		{onChange}
		disabled={onMultiDiscover}
	/>
</div>
