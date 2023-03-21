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

import { LOCAL_STORAGE_HTTP_PORT, LOCAL_STORAGE_WS_PORT } from './const/localStorageConsts';

const ws_port = localStorage.getItem(LOCAL_STORAGE_WS_PORT);
const http_port = localStorage.getItem(LOCAL_STORAGE_HTTP_PORT);
const environment = process.env.NODE_ENV ? process.env.NODE_ENV : 'development';
const SERVER_URL_PRODUCTION = `${window.location.href.split('//')[1].split('/')[0]}/api`;
var WS_SERVER_URL_PRODUCTION = `${window.location.href.split('//')[1].split('/')[0]}`;
if (WS_SERVER_URL_PRODUCTION.includes(':'))
    // for urls contain port
    WS_SERVER_URL_PRODUCTION = WS_SERVER_URL_PRODUCTION.split(':')[0];
const SSL_PREFIX = window.location.protocol === 'https:' ? 'https://' : 'http://';

export const SERVER_URL = environment === 'production' ? `${SSL_PREFIX}${SERVER_URL_PRODUCTION}` : `http://localhost:${http_port}/api`;
const WS_PREFIX = window.location.href.includes('https') ? 'wss' : 'ws';
export const URL = window.location.href;

export const HANDLE_REFRESH_INTERVAL = 600000;
export const SHOWABLE_ERROR_STATUS_CODE = 666;
export const SANDBOX_SHOWABLE_ERROR_STATUS_CODE = 665;
export const AUTHENTICATION_ERROR_STATUS_CODE = 401;
export const DOC_URL = 'https://docs.memphis.dev/memphis/memphis/overview';
export const SOCKET_URL = environment === 'production' ? `${WS_PREFIX}://${WS_SERVER_URL_PRODUCTION}:${ws_port}` : `${WS_PREFIX}://localhost:${ws_port}`;
export const GOOGLE_CLIENT_ID = '916272522459-u0f4n2lh9llsielb3l5rob3dnt1fco76.apps.googleusercontent.com';
export const REDIRECT_URI = environment === 'production' ? 'https://sandbox.memphis.dev/login' : `http://localhost:${http_port}/login`;
export const GITHUB_CLIENT_ID = environment === 'production' ? '4dc1b3238c4d7563e426' : '51b0330eb3b34bc8f641';
export const CONNECT_APP_VIDEO = 'https://www.youtube.com/watch?v=-5YmxYRQsdw';
export const CONNECT_CLI_VIDEO = 'https://www.youtube.com/watch?v=awXwaU4rBBQ';

export const RELEASE_NOTES_URL = 'https://api.github.com/repos/Memphisdev/gitbook-backup/contents/release-notes/releases';
export const LATEST_RELEASE_URL = 'https://api.github.com/repos/Memphisdev/memphis/releases';
export const RELEASE_DOCS_URL = 'https://docs.memphis.dev/memphis/release-notes/releases/';
export const DOCKER_UPGRADE_URL = 'https://docs.memphis.dev/memphis/deployment/docker-compose#how-to-upgrade';
export const K8S_UPGRADE_URL = 'https://docs.memphis.dev/memphis/deployment/kubernetes/how-to-upgrade';
