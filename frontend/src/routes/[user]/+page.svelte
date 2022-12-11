<script lang="ts">
    import Posts from "../../components/post/Posts.svelte";
    import {
        addNewUserPost,
        userNewPostsStore,
        userPostsStore
    } from "../../actions/posts";
    import {onDestroy, onMount} from "svelte";
    import {createPost, registerPostsUpdate} from "../../services/posts";
    import type PostData from "../../types/PostData";
    import {env} from "$env/dynamic/public";
    import Fab from "../../components/Fab.svelte";
    import {closeNewPostModal, openNewPostModal} from "../../actions/newPostModal.js";
    import NewPostModal from "../../components/post/NewPostModal.svelte";
    import UserBadge from "../../components/users/UserBadge.svelte";
    import {followUser} from "../../services/users";
    import {refreshUserTimeline} from "../../actions/posts.js";

    export let data;

    let posts: PostData[]
    userPostsStore.subscribe((value) => posts = value)

    let newPosts: PostData[]
    userNewPostsStore.subscribe((value) => newPosts = value)

    onMount(() => {
       const sse = registerPostsUpdate((post) => {
            post.username === data.user.username && addNewUserPost(post)
        });

       return () => sse.close()
    })
    let loading;

    let handleCreatePost = async (data) => {
        loading = true
        try {
            const post = {
                username: env.PUBLIC_USERNAME,
                text: data.content,
                timestamp: new Date()
            };

            await createPost(post);
            addNewUserPost(post)

            closeNewPostModal()
            loading = false
            return true
        } catch (e) {
            console.error(e)
            loading = false
            return true
        }
    }

    const submitFollow = async () => {
        loading = true
        try {
            const target = !data.user.following
            await followUser(data.user.username, target);
            data.user = {...data.user, following: target}
            loading = false;
            return true;
        } catch (e) {
            console.error(e)
            loading = false;
            return false;
        }
    }

</script>

{#if data.error}
    <div class="flex items-center justify-center m-5">
        <span class="text-xl text-error">
            The user you are trying to access does not exist
        </span>
    </div>
{:else}

    <div class="grid sm:grid-cols-4 gap-4 sm:mx-20 mt-0">
        <div class="sm:col-span-1 col-span-4">
            <div class="sticky top-10">
                <UserBadge
                    user={data.user}
                    margin="my-[5em] mx-0"
                    background="bg-base-100"
                    loading={loading}
                    follow={submitFollow}
                />

            </div>
        </div>
        <div class="sm:col-span-3 col-span-4">
            <Posts refreshTimeline={refreshUserTimeline}
                   emptyPostsMessage="This user has not posted anything"
                   posts={posts}
                   newPosts={newPosts}
                   mainTimeline={false}/>
        </div>
    </div>


    {#if data.user.username === env.PUBLIC_USERNAME}
        <Fab action={openNewPostModal}/>

        <NewPostModal loading={loading} close={closeNewPostModal} submit={handleCreatePost}/>
    {/if}

{/if}