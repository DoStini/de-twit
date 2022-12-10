import {writable} from "svelte/store";
import type PostData from "../types/PostData";

export const postsStore = writable<PostData[]>([])
export const newPostsStore = writable<PostData[]>([])

export const addPost = (post: PostData) => postsStore.update((posts) => [post, ...posts])
export const setPosts = (posts: PostData[]) => postsStore.set(posts)
