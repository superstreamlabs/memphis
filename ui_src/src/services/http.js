// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

import axios from 'axios';

import { SERVER_URL, SHOWABLE_ERROR_STATUS_CODE, AUTHENTICATION_ERROR_STATUS_CODE, CLOUD_URL } from '../config';
import { LOCAL_STORAGE_TOKEN } from '../const/localStorageConsts.js';
import { ApiEndpoints } from '../const/apiEndpoints';
import pathDomains from '../router';
import AuthService from './auth';
import { isCloud } from './valueConvertor';
import EmailLink from '../components/emailLink';
import { showMessages } from './genericServices';

export async function httpRequest(method, endPointUrl, data = {}, headers = {}, queryParams = {}, authNeeded = true, timeout = 0, serverUrl = null, displayMsg = true) {
    if (authNeeded) {
        const token = localStorage.getItem(LOCAL_STORAGE_TOKEN);
        headers['Authorization'] = 'Bearer ' + token;
    }

    const HTTP = axios.create({
        withCredentials: serverUrl ? false : true
    });
    if (method !== 'GET' && method !== 'POST' && method !== 'PUT' && method !== 'DELETE')
        throw {
            status: 400,
            message: `Invalid HTTP method`,
            data: { method, endPointUrl, data }
        };
    try {
        const url = `${serverUrl || SERVER_URL}${endPointUrl}`;
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
            err?.response?.status === AUTHENTICATION_ERROR_STATUS_CODE &&
            !serverUrl
        ) {
            AuthService.clearLocalStorage();
            isCloud() ? window.location.replace(CLOUD_URL) : window.location.assign(pathDomains.login);
        }
        if (err?.response?.data?.message !== undefined && err?.response?.status === SHOWABLE_ERROR_STATUS_CODE && displayMsg) {
            showMessages('warning', err?.response?.data?.message);
        }
        if (err?.response?.data?.message !== undefined && err?.response?.status === 500) {
            showMessages(
                'error',
                isCloud() ? (
                    <>
                        We are experiencing some issues. Please contact us at <EmailLink email="support@memphis.dev" /> for assistance.
                    </>
                ) : (
                    <>
                        <>
                            We have some issues. Please open a
                            <a className="a-link" href="https://github.com/memphisdev/memphis" target="_blank">
                                GitHub issue
                            </a>
                        </>
                    </>
                )
            );
        }
        if (err?.message?.includes('Network Error') && serverUrl) {
            showMessages('warning', `${serverUrl} can not be reached`);
        }
        throw err.response;
    }
}

export async function handleRefreshTokenRequest() {
    const HTTP = axios.create({
        withCredentials: true
    });
    try {
        const url = `${SERVER_URL}${ApiEndpoints.REFRESH_TOKEN}`;
        const res = await HTTP({ method: 'POST', url });
        await AuthService.saveToLocalStorage(res.data);
        return res.data;
    } catch (err) {
        AuthService.clearLocalStorage();
        isCloud() ? window.location.replace(CLOUD_URL) : window.location.assign(pathDomains.login);
        return '';
    }
}
