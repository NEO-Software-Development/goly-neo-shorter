<script>
	/**
	 * @component
	 * This component fetches and displays a list of Goly links.
	 * It uses the Item component to render each link.
	 */
    import Item from "./Item.svelte"
    import { onMount } from "svelte"

	/**
	 * An array of Goly objects.
	 * @type {Array<object>}
	 */
    let golies = []

    onMount(async () => {
        const res = await fetch("http://localhost:3000/goly")
        golies = await res.json()
    })
</script>

{#each golies as goly}
    <Item goly={goly} />
{/each}
