<script lang="ts">

    import type UserData from "../../types/UserData";
    import {env} from "$env/dynamic/public";

    export let user: UserData
    export let error: boolean
    export let loading: boolean
    export let background: string;
    export let margin: string
    export let redirect: boolean;

    export let follow: () => (Promise<boolean>)
    export let close


    const onFollow = async () => {
        await follow()
    }

</script>

<div class="card {background ? background : 'bg-base-200'} shadow-xl {margin ? margin :  'my-5 mx-0'}">
    <div class="card-body">
        {#if error}
            <div class="text-xl">User not found</div>
        {:else }
            <div class="flex justify-between mb-4 content-center">
                <div>
                    <div class="text-xl text-accent"><a on:click={() => close && close()} href="/{user.username}">{user.username}</a></div>
                    <div class="text-sm">
                        {#if user.posts.length === 0}
                            This user hasn't posted anything yet
                        {:else if user.posts.length === 1}
                            1 Post
                        {:else}
                            {user.posts.length.toString()} Posts
                        {/if}
                    </div>
                </div>
                <div class="mask mask-hexagon h-14 w-14">
                    <img src="https://placeimg.com/192/192/people" />
                </div>
            </div>

            {#if user.username !== env.PUBLIC_USERNAME}
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
        {/if}
    </div>
</div>
