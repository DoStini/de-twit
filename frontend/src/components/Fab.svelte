<script>
    import PlusIcon from "../icons/PlusIcon.svelte";

    export let action;
</script>
<button on:click={action} class="btn btn-circle btn-primary fixed bottom-5 right-5" >
    <PlusIcon/>
</button>
