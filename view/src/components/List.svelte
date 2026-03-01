<script>
    import Item from "./Item.svelte"
    import { onMount } from "svelte"
    import { golies } from "../stores.js"

    onMount(async () => {
        const res = await fetch("http://localhost:3000/goly")
        const data = await res.json()
        golies.set(data)
    })
</script>

{#each $golies as goly (goly.id)}
    <Item goly={goly} />
{/each}
