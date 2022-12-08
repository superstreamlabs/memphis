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

import { message } from 'antd';
import axios from 'axios';

import { SERVER_URL, SHOWABLE_ERROR_STATUS_CODE, AUTHENTICATION_ERROR_STATUS_CODE } from '../config';
import { ApiEndpoints } from '../const/apiEndpoints';
import { LOCAL_STORAGE_TOKEN, LOCAL_STORAGE_EXPIRED_TOKEN, LOCAL_STORAGE_SKIP_GET_STARTED } from '../const/localStorageConsts.js';
import pathDomains from '../router';
import AuthService from './auth';

export async function httpRequest(method, endPointUrl, data = {}, headers = {}, queryParams = {}, authNeeded = true, timeout = 0) {
    let isSkipGetStarted;
    if (authNeeded) {
        const token = localStorage.getItem(LOCAL_STORAGE_TOKEN);
        headers['Authorization'] = 'Bearer ' + token;
    }
    const HTTP = axios.create({
        withCredentials: true
    });
    if (method !== 'GET' && method !== 'POST' && method !== 'PUT' && method !== 'DELETE')
        throw {
            status: 400,
            message: `Invalid HTTP method`,
            data: { method, endPointUrl, data }
        };
    try {
        const url = `${SERVER_URL}${endPointUrl}`;
        const res = await HTTP({
            method,
            url,
            data,
            headers,
            timeout,
            params: queryParams
        });
        const results = res.data;
        return results;
    } catch (err) {
        if (
            window.location.pathname !== pathDomains.login &&
            window.location.pathname !== pathDomains.signup &&
            err?.response?.status === AUTHENTICATION_ERROR_STATUS_CODE
        ) {
            isSkipGetStarted = localStorage.getItem(LOCAL_STORAGE_SKIP_GET_STARTED);
            localStorage.clear();
            if (isSkipGetStarted === 'true') {
                localStorage.setItem(LOCAL_STORAGE_SKIP_GET_STARTED, isSkipGetStarted);
            }
            window.location.assign('/login');
        }
        if (err?.response?.data?.message !== undefined && err?.response?.status === SHOWABLE_ERROR_STATUS_CODE) {
            message.warning({
                key: 'memphisWarningMessage',
                content: err?.response?.data?.message,
                duration: 5,
                style: { cursor: 'pointer' },
                onClick: () => message.destroy('memphisWarningMessage')
            });
        }
        if (err?.response?.data?.message !== undefined && err?.response?.status === 500) {
            message.error({
                key: 'memphisErrorMessage',
                content: (
                    <>
                        We have some issues. Please open a
                        <a className="a-link" href="https://github.com/memphisdev/memphis-broker" target="_blank">
                            GitHub issue
                        </a>
                    </>
                ),
                duration: 5,
                style: { cursor: 'pointer' },
                onClick: () => message.destroy('memphisErrorMessage')
            });
        }
        throw err.response;
    }
}

export async function handleRefreshTokenRequest() {
    let isSkipGetStarted;
    const HTTP = axios.create({
        withCredentials: true
    });
    try {
        const url = `${SERVER_URL}${ApiEndpoints.REFRESH_TOKEN}`;
        const res = await HTTP({ method: 'POST', url });
        const now = new Date();
        const expiryToken = now.getTime() + res.data.expires_in;
        if (process.env.REACT_APP_SANDBOX_ENV) {
            localStorage.setItem(LOCAL_STORAGE_TOKEN, res.data.jwt);
            localStorage.setItem(LOCAL_STORAGE_EXPIRED_TOKEN, expiryToken);
        } else {
            await AuthService.saveToLocalStorage(res.data);
        }
        return true;
    } catch (err) {
        isSkipGetStarted = localStorage.getItem(LOCAL_STORAGE_SKIP_GET_STARTED);
        localStorage.clear();
        if (isSkipGetStarted === 'true') {
            localStorage.setItem(LOCAL_STORAGE_SKIP_GET_STARTED, isSkipGetStarted);
        }
        window.location.assign('/login');
        return false;
    }
}
