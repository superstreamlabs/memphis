// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

export const ApiEndpoints = {
    //Auth
    LOGIN: '/usermgmt/login',
    SIGNUP: '/usermgmt/addUserSignUp',
    REFRESH_TOKEN: '/usermgmt/refreshToken',
    ADD_USER: '/usermgmt/addUser',
    GET_ALL_USERS: '/usermgmt/getAllUsers',
    REMOVE_USER: '/usermgmt/removeUser',
    REMOVE_MY_UER: '/usermgmt/removeMyUser',
    EDIT_AVATAR: '/usermgmt/editAvatar',
    GET_COMPANY_LOGO: '/usermgmt/getCompanyLogo',
    EDIT_COMPANY_LOGO: '/usermgmt/editCompanyLogo',
    REMOVE_COMPANY_LOGO: '/usermgmt/removeCompanyLogo',
    EDIT_ANALYTICS: '/usermgmt/editAnalytics',
    SANDBOX_LOGIN: '/sandbox/login',
    DONE_NEXT_STEPS: '/usermgmt/doneNextSteps',
    GET_SIGNUP_FLAG: '/usermgmt/getSignUpFlag',

    //Station
    CREATE_STATION: '/stations/createStation',
    REMOVE_STATION: '/stations/removeStation',
    GET_STATION: '/stations/getStation',
    GET_ALL_STATIONS: '/stations/getAllStations',
    GET_STATIONS: '/stations/getStations',
    GET_POISION_MESSAGE_JOURNEY: '/stations/getPoisonMessageJourney',
    GET_MESSAGE_DETAILS: '/stations/getMessageDetails',
    ACK_POISION_MESSAGE: '/stations/ackPoisonMessages',
    RESEND_POISION_MESSAGE_JOURNEY: '/stations/resendPoisonMessages',
    USE_SCHEMA: '/stations/useSchema',
    GET_UPDATE_SCHEMA: '/stations/getUpdatesForSchemaByStation',

    //Producers
    GET_ALL_PRODUCERS_BY_STATION: '/producers/getAllProducersByStation',

    //Consumers
    GET_ALL_CONSUMERS_BY_STATION: '/consumers/getAllConsumersByStation',

    //Monitor
    GET_CLUSTER_INFO: '/monitoring/getClusterInfo',
    GET_MAIN_OVERVIEW_DATA: '/monitoring/getMainOverviewData',
    GET_STATION_DATA: '/monitoring/getStationOverviewData',
    GET_SYS_LOGS: '/monitoring/getSystemLogs',
    DOWNLOAD_SYS_LOGS: '/monitoring/downloadSystemLogs',

    //Tags
    GET_TAGS: '/tags/getTags',
    GET_USED_TAGS: '/tags/getUsedTags',
    REMOVE_TAG: '/tags/removeTag',
    CREATE_NEW_TAG: '/tags/createNewTag',
    UPDATE_TAGS_FOR_ENTITY: '/tags/updateTagsForEntity',

    //Schemas
    GET_ALL_SCHEMAS: '/schemas/getAllSchemas',
    CREATE_NEW_SCHEMA: '/schemas/createNewSchema',
    GET_SCHEMA_DETAILS: '/schemas/getSchemaDetails',
    REMOVE_SCHEMA: '/schemas/removeSchema',
    CREATE_NEW_VERSION: '/schemas/createNewVersion',
    ROLL_BACK_VERSION: '/schemas/rollBackVersion',
    VALIDATE_SCHEMA: '/schemas/validateSchema'
};
