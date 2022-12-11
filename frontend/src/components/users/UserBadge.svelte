<script lang="ts">

    import type UserData from "../../types/UserData";

    export let user: UserData
    export let error: boolean
    export let loading: boolean

    export let follow: () => (Promise<boolean>)

    const onFollow = async () => {
        await follow()
    }

</script>

<div class="card bg-base-200 shadow-xl my-5 mx-0">
    <div class="card-body">
        {#if error}
            <div class="text-xl">User not found</div>
        {:else }
            <div class="text-xl text-accent">{user.username}</div>
            <div class="text-sm">{user.posts.length.toString()} Posts</div>
            <div class="card-actions justify-end">
                {#if !user.following}
                    <div class="btn text-success {loading ? 'loading' : ''}" on:click={onFollow}>
                        Follow
                    </div>
                {:else }
                    <div class="btn text-error {loading ? 'loading' : ''}" on:click={onFollow}>
                        Unfollow
                    </div>
                {/if}
            </div>
        {/if}
    </div>
</div>
