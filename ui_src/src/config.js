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

export const ENVIRONMENT = process.env.NODE_ENV ? process.env.NODE_ENV : 'development';
const SERVER_URL_PRODUCTION = `${window.location.href.split('//')[1].split('/')[0]}/api`;
const SSL_PREFIX = window.location.protocol === 'https:' ? 'https://' : 'http://';

export const SERVER_URL = ENVIRONMENT === 'production' ? `${SSL_PREFIX}${SERVER_URL_PRODUCTION}` : `http://localhost:9000/api`;
export const URL = window.location.href;

export const HANDLE_REFRESH_INTERVAL = 600000;
export const SHOWABLE_ERROR_STATUS_CODE = 666;
export const AUTHENTICATION_ERROR_STATUS_CODE = 401;
export const DOC_URL = 'https://docs.memphis.dev/memphis/getting-started/readme';
export const CONNECT_APP_VIDEO = 'https://www.youtube.com/watch?v=-5YmxYRQsdw';
export const CONNECT_CLI_VIDEO = 'https://www.youtube.com/watch?v=awXwaU4rBBQ';
export const WELCOME_VIDEO = 'https://youtu.be/mDbGnMC1hig';
export const FUNCTION_GUIDE_VIDEO = 'https://www.youtube.com/watch?v=-5YmxYRQsdw';
export const RELEASE_NOTES_URL = 'https://api.github.com/repos/Memphisdev/gitbook-backup/contents/release-notes/releases';
export const LATEST_RELEASE_URL = 'https://api.github.com/repos/Memphisdev/memphis/releases';
export const RELEASE_DOCS_URL = 'https://docs.memphis.dev/memphis/release-notes/releases/';
export const DOCKER_UPGRADE_URL = 'https://docs.memphis.dev/memphis/deployment/docker-compose#how-to-upgrade';
export const K8S_UPGRADE_URL = 'https://docs.memphis.dev/memphis/deployment/kubernetes/how-to-upgrade';

export const CLOUD_URL =
    process.env.REACT_APP_CLOUD === 'true'
        ? window.location.href?.includes('localhost')
            ? 'http://localhost:10005/signin'
            : window.location.href?.includes('cloud-qa')
            ? 'https://cloud-qa.memphis.dev/signin'
            : window.location.href?.includes('cloud-staging')
            ? 'https://cloud-staging.memphis.dev/signin'
            : 'https://cloud.memphis.dev'
        : null;

export const STIGG_CLIENT_ID =
    process.env.REACT_APP_CLOUD === 'true'
        ? window.location.href?.includes('localhost')
            ? 'client-becd87e7-8dba-4c0a-8e9f-bb10d5b0a1ff:04308cac-952e-4f75-a503-b3d93be85c9a'
            : window.location.href?.includes('cloud-qa')
            ? 'client-0abc22d6-1a09-40d4-bed8-2b682df63149:04308cac-952e-4f75-a503-b3d93be85c9a'
            : window.location.href?.includes('cloud-staging')
            ? 'client-05a500b6-81d3-4af5-afd9-8ea72f727edf:04308cac-952e-4f75-a503-b3d93be85c9a'
            : 'client-da938764-f7b9-406b-8d02-d2f834a795b4:04308cac-952e-4f75-a503-b3d93be85c9a'
        : null;
