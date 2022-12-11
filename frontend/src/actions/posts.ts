import {writable} from "svelte/store";
import type PostData from "../types/PostData";

export const postsStore = writable<PostData[]>([])
export const newPostsStore = writable<PostData[]>([])

export const userPostsStore = writable<PostData[]>([])
export const userNewPostsStore = writable<PostData[]>([])


export const addPost = (post: PostData) => postsStore.update((posts) => [post, ...posts])
export const setPosts = (posts: PostData[]) => postsStore.set(posts)

export const addNewPost = (post: PostData) => newPostsStore.update((posts) => [post, ...posts])

export const refreshTimeline = () => {
    let newPosts: PostData[];

    newPostsStore.update((posts) => {
        newPosts = posts;
        return [];
    })

    postsStore.update((posts) => [...newPosts, ...posts]);
}

export const addNewUserPost = (post: PostData) => userNewPostsStore.update((posts) => [post, ...posts])

export const refreshUserTimeline = () => {
    let newPosts: PostData[];

    userNewPostsStore.update((posts) => {
        newPosts = posts;
        return [];
    })

    userPostsStore.update((posts) => [...newPosts, ...posts]);
}

