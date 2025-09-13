<script>
	/**
	 * @component
	 * This component provides a button to open a modal for creating a new Goly link.
	 */
    import { Modals, closeModal, openModal } from "svelte-modals"
    import Modal from "./Modal.svelte"

	/**
	 * Creates a new Goly link on the server.
	 * @param {{redirect: string, goly: string, random: boolean}} data - The data for the new Goly link.
	 */
    async function updateGoly(data) {
        const json = {
            redirect: data.redirect,
            goly: data.goly,
            random: data.random
        }
        await fetch("http://localhost:3000/goly", {
            method: "POST",
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify(json)
        }).then(response => {
            console.log(response)
        })
    }

	/**
	 * Opens the modal for creating a new Goly link.
	 */
    function handleOpen() {
        openModal(Modal, {
            title: "Create New Goly Link",
            send: updateGoly,
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