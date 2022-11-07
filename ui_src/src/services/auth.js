// Credit for The NATS.IO Authors
// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package server

import {
    LOCAL_STORAGE_ALREADY_LOGGED_IN,
    LOCAL_STORAGE_AVATAR_ID,
    LOCAL_STORAGE_CREATION_DATE,
    LOCAL_STORAGE_TOKEN,
    LOCAL_STORAGE_EXPIRED_TOKEN,
    LOCAL_STORAGE_USER_ID,
    LOCAL_STORAGE_USER_NAME,
    LOCAL_STORAGE_USER_TYPE,
    LOCAL_STORAGE_ALLOW_ANALYTICS,
    LOCAL_STORAGE_ENV,
    LOCAL_STORAGE_NAMESPACE,
    LOCAL_STORAGE_WELCOME_MESSAGE,
    LOCAL_STORAGE_FULL_NAME,
    LOCAL_STORAGE_SKIP_GET_STARTED
} from '../const/localStorageConsts';
import pathDomains from '../router';

const AuthService = (function () {
    const saveToLocalStorage = (userData) => {
        const now = new Date();
        const expiryToken = now.getTime() + userData.expires_in;

        localStorage.setItem(LOCAL_STORAGE_ALREADY_LOGGED_IN, userData.already_logged_in);
        localStorage.setItem(LOCAL_STORAGE_AVATAR_ID, userData.avatar_id);
        localStorage.setItem(LOCAL_STORAGE_CREATION_DATE, userData.creation_date);
        localStorage.setItem(LOCAL_STORAGE_TOKEN, userData.jwt);
        localStorage.setItem(LOCAL_STORAGE_USER_ID, userData.user_id);
        localStorage.setItem(LOCAL_STORAGE_USER_NAME, userData.username);
        localStorage.setItem(LOCAL_STORAGE_FULL_NAME, userData.full_name);
        localStorage.setItem(LOCAL_STORAGE_USER_TYPE, userData.user_type);
        localStorage.setItem(LOCAL_STORAGE_EXPIRED_TOKEN, expiryToken);
        localStorage.setItem(LOCAL_STORAGE_ALLOW_ANALYTICS, userData.send_analytics);
        localStorage.setItem(LOCAL_STORAGE_ENV, userData.env);
        localStorage.setItem(LOCAL_STORAGE_NAMESPACE, userData.namespace);
        if (userData.already_logged_in === false) {
            localStorage.setItem(LOCAL_STORAGE_WELCOME_MESSAGE, true);
        }
    };

    const logout = () => {
        const isSkipGetStarted = localStorage.getItem(LOCAL_STORAGE_SKIP_GET_STARTED);
        localStorage.clear();
        if (isSkipGetStarted === 'true') {
            localStorage.setItem(LOCAL_STORAGE_SKIP_GET_STARTED, isSkipGetStarted);
        }
        window.location.assign(pathDomains.login);
    };

    const isValidToken = () => {
        const tokenExpiryTime = localStorage.getItem(LOCAL_STORAGE_EXPIRED_TOKEN);
        if (Date.now() <= tokenExpiryTime) {
            return true;
        } else {
            return false;
        }
    };

    return {
        saveToLocalStorage,
        logout,
        isValidToken
    };
})();
export default AuthService;
