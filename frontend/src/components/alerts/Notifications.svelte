<script lang="ts">
    import type NotificationData from "../../types/NotificationData";
    import {notificationsStore} from "../../actions/notifications";
    import ErrorAlert from "./ErrorAlert.svelte";
    import type NotificationRecord from "../../types/NotificationRecord";

    let notifications: NotificationRecord;
    notificationsStore.subscribe((value) => notifications = value);

    console.log(notifications)

</script>

<div class="fixed w-full p-10">
    {#each Object.entries(notifications).sort(([_1, v1], [_2, v2]) => v2.timestamp - v1.timestamp) as [id, notification]}
        <div class="mb-5">
            <ErrorAlert id={id} type="{notification.type}" text={notification.text}></ErrorAlert>
        </div>
    {/each}
</div>
