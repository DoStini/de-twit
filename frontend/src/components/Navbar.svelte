<script lang="ts">
import {openSearchUserModal} from "../actions/searchUserModal.js";
import {env} from "$env/dynamic/public";
</script>

<div class="navbar bg-base-100">
    <div class="navbar-start">

    </div>
    <div class="navbar-center">
        <a href="/" class="btn btn-ghost normal-case text-xl">De-Twit</a>
    </div>
    <div class="navbar-end">
        <button class="btn btn-ghost btn-circle" on:click={openSearchUserModal}>
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
        </button>
        <button class="btn btn-ghost btn-circle">
            <div class="indicator">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" /></svg>
                <span class="badge badge-xs badge-primary indicator-item"></span>
            </div>
        </button>

        <button class="dropdown dropdown-end">
            <label class="btn btn-ghost">
                <div class="avatar">
                    <div class="w-8 mask mask-hexagon">
                        <img src="https://placeimg.com/192/192/people" />
                    </div>
                </div>
            </label>
            <ul tabindex="0" class="dropdown-content menu p-2 shadow bg-base-100 rounded-box w-52">
                <li><a href="/{env.PUBLIC_USERNAME}">View profile</a></li>
                <li><a class="text-error">Delete account</a></li>
            </ul>
        </button>
    </div>
</div>
