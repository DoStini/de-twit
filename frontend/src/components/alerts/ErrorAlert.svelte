<script lang="ts">
    import CrossIcon from "../../icons/CrossIcon.svelte";
    import {notificationsStore} from "../../actions/notifications";
    import {closeNotification} from "../../actions/notifications.js";

    export let id;
    export let text;
    export let type;

    let show: boolean
    notificationsStore.subscribe((value) => (console.log(value[id].open), show = value[id].open))

</script>

<div class="alert alert-{type} z-{show ? '20' : '-100'} transition-all shadow-lg opacity-{show ? '10' : '0'}">
    <div>
        <button class="btn btn-ghost" on:click={() => closeNotification(id)}>
            <CrossIcon/>
        </button>
        <span>Error! Task failed successfully.</span>
    </div>
</div>
