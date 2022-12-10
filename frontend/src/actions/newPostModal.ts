import {writable} from "svelte/store";

export const newPostModalStore = writable<boolean>(false);

export const openNewPostModal = () => newPostModalStore.set(true);
export const closeNewPostModal = () => newPostModalStore.set(false);
