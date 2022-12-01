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
    SKIP_GET_STARTED: '/usermgmt/skipGetStarted',

    //Station
    CREATE_STATION: '/stations/createStation',
    REMOVE_STATION: '/stations/removeStation',
    GET_STATION: '/stations/getStation',
    GET_ALL_STATIONS: '/stations/getAllStations',
    GET_STATIONS: '/stations/getStations',
    GET_POISON_MESSAGE_JOURNEY: '/stations/getPoisonMessageJourney',
    GET_MESSAGE_DETAILS: '/stations/getMessageDetails',
    ACK_POISON_MESSAGE: '/stations/ackPoisonMessages',
    RESEND_POISON_MESSAGE_JOURNEY: '/stations/resendPoisonMessages',
    USE_SCHEMA: '/stations/useSchema',
    GET_UPDATE_SCHEMA: '/stations/getUpdatesForSchemaByStation',
    REMOVE_SCHEMA_FROM_STATION: '/stations/removeSchemaFromStation',
    TIERD_STORAGE_CLICKED: '/stations/tierdStorageClicked',

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
    VALIDATE_SCHEMA: '/schemas/validateSchema',

    //Integrations
    CREATE_INTEGRATION: '/integrations/createIntegration',
    UPDATE_INTEGRATIONL: '/integrations/updateIntegration',
    GET_INTEGRATION_DETAILS: '/integrations/getIntegrationDetails',
    GET_ALL_INTEGRATION: '/integrations/getAllIntegrations',
    DISCONNECT_INTEGRATION: '/integrations/disconnectIntegration',
    REQUEST_INTEGRATION: '/integrations/requestIntegration'
};
