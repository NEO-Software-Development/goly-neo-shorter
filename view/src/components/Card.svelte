<script>
    /**
     * @component
     * A simple card component for displaying content.
     * It provides a styled container for its children.
     */
</script>

<div class="card">
    <slot></slot>
</div>
<style>
    .card {
        background-color: steelblue;
        color: white;
        border-radius: 15px;
        padding: 50px;
        margin: 20px 10px;
        max-width: 800px;
    }
</style>