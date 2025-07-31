<script>
    import { Modals, closeModal, openModal } from "svelte-modals"
    import Modal from "./Modal.svelte"
    import { golies } from "../stores.js"

    async function createGoly(data) {
        const json = {
            redirect: data.redirect,
            goly: data.goly,
            random: data.random
        }
        const res = await fetch("http://localhost:3000/goly", {
            method: "POST",
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify(json)
        })
        const newGoly = await res.json()
        golies.update(items => [newGoly, ...items])
        closeModal()
    }

    function handleOpen() {
        openModal(Modal, {
            title: "Create New Goly Link",
            send: createGoly,
            redirect: "",
            goly: "",
            random: false
        })
    }
</script>

<button on:click={ handleOpen }>New</button>

<style>
    button {
        background-color: green;
        color: white;
        font-weight: bold;
        border: none;
        padding: .75rem;
        border-radius: 4px;
    }
</style>