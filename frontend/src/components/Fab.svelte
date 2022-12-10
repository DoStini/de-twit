<script>
    import PlusIcon from "../icons/PlusIcon.svelte";
    import {launchNotification} from "../actions/notifications";

    export let action;
</script>
<button on:click={action} class="btn btn-circle btn-primary fixed bottom-5 right-5" >
    <PlusIcon/>
</button>
