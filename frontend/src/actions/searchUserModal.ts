import {writable} from "svelte/store";

export const searchUserModal = writable<boolean>(false);

export const openSearchUserModal = () => searchUserModal.set(true);
export const closeSearchUserModal = () => searchUserModal.set(false);
