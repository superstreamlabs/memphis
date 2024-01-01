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

import './style.scss';

import React, { useEffect, useState, useRef } from 'react';
import { FiMinusCircle, FiPlus } from 'react-icons/fi';
import Editor, { loader } from '@monaco-editor/react';
import { Divider, Form, Collapse } from 'antd';
import * as monaco from 'monaco-editor';

import { REST_CODE_EXAMPLE, SDK_CODE_EXAMPLE, sdkLangOptions, restLangOptions } from '../../const/codeExample';
import {
    LOCAL_STORAGE_ACCOUNT_ID,
    LOCAL_STORAGE_BROKER_HOST,
    LOCAL_STORAGE_ENV,
    LOCAL_STORAGE_REST_GW_HOST,
    LOCAL_STORAGE_REST_GW_PORT,
    LOCAL_STORAGE_USER_PASS_BASED_AUTH,
    LOCAL_STORAGE_PRODUCER_COMMUNICATION_TYPE,
    LOCAL_STORAGE_PRODUCER_PROGRAMMING_LANGUAGE,
    LOCAL_STORAGE_CONSUMER_COMMUNICATION_TYPE,
    LOCAL_STORAGE_CONSUMER_PROGRAMMING_LANGUAGE,
} from '../../const/localStorageConsts';
import GenerateTokenModal from '../../domain/stationOverview/components/generateTokenModal';
import { ReactComponent as NoCodeExampleIcon } from '../../assets/images/noCodeExample.svg';
import { ReactComponent as RefreshIcon } from '../../assets/images/refresh.svg';
import { ReactComponent as CodeIcon } from '../../assets/images/codeIcon.svg';
import CreateUserDetails from '../../domain/users/createUserDetails';
import CollapseArrow from '../../assets/images/collapseArrow.svg';
import { ApiEndpoints } from '../../const/apiEndpoints';
import TitleComponent from '../titleComponent/index';
import { httpRequest } from '../../services/http';
import SegmentButton from '../segmentButton';
import CustomSelect from '../customSelect';
import SelectComponent from '../select';
import Switcher from '../switcher';
import Modal from '../modal';
import Input from '../Input';
import Copy from '../copy';
import { Drawer } from 'antd';

loader.init();
loader.config({ monaco });

const tabs = ['Producer', 'Consumer'];
const tabsProtocol = ['Generate token', 'Produce data', 'Consume data'];
const selectProtocolOption = ['SDK', 'REST'];

const SdkExample = ({ consumer, showTabs = true, stationName, username, connectionCreds, withHeader = false }) => {
    const [langSelected, setLangSelected] = useState('Go');
    const [protocolSelected, setProtocolSelected] = useState('SDK');
    const [codeExample, setCodeExample] = useState({
        producer: '',
        consumer: ''
    });
    const [tabValue, setTabValue] = useState(consumer ? 'Consumer' : 'Producer');
    const [tabValueRest, setTabValueRest] = useState('Generate token');
    const [generateModal, setGenerateModal] = useState(false);
    const [formFields, setFormFields] = useState({
        userName: '',
        password: '',
        producerConsumerName: '',
        consumerGroupName: '',
        blocking: true,
        async: true,
        useHeaders: true,
        headersList: [{ key: '', value: '' }],
        jwt: '',
        tokenExpiry: 100,
        refreshToken: 100,
        batchSize: 10,
        batchTimeout: 1000
    });
    const [createUserLoader, setCreateUserLoader] = useState(false);
    const [addUserModalIsOpen, addUserModalFlip] = useState(false);
    const [users, setUsers] = useState([]);
    const createUserRef = useRef(null);
    const { Panel } = Collapse;

    const getAllUsers = async () => {
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_ALL_USERS);
            setUsers(data.application_users.map((user) => ({ ...user, name: user.username })));
        } catch (error) {}
    };

    useEffect(() => {
        if (!consumer) {
            getAllUsers();
        }
    }, []);

    useEffect(() => {
        protocolSelected === 'SDK' ? changeSDKDynamicCode(langSelected) : changeRestDynamicCode(langSelected);
    }, []);

    useEffect(() => {
        protocolSelected === 'SDK' ? changeSDKDynamicCode(langSelected) : changeRestDynamicCode(langSelected);
    }, [formFields, tabValue]);

    useEffect(() => {
        const communicationTypeStorage = !consumer ? LOCAL_STORAGE_PRODUCER_COMMUNICATION_TYPE : LOCAL_STORAGE_CONSUMER_COMMUNICATION_TYPE;
        const communicationType = localStorage.getItem(communicationTypeStorage);
        if (communicationType) setProtocolSelected(communicationType);

        const programmingLanguageStorage = !consumer ? LOCAL_STORAGE_PRODUCER_PROGRAMMING_LANGUAGE : LOCAL_STORAGE_CONSUMER_PROGRAMMING_LANGUAGE;
        const programmingLanguage = localStorage.getItem(programmingLanguageStorage);
        if (programmingLanguage) setLangSelected(programmingLanguage);
    }, []);

    useEffect(() => {
        const communicationType = !consumer ? LOCAL_STORAGE_PRODUCER_COMMUNICATION_TYPE : LOCAL_STORAGE_CONSUMER_COMMUNICATION_TYPE;
        localStorage.setItem(communicationType, protocolSelected);
    }, [protocolSelected]);

    useEffect(() => {
        const programmingLanguage = !consumer ? LOCAL_STORAGE_PRODUCER_PROGRAMMING_LANGUAGE : LOCAL_STORAGE_CONSUMER_PROGRAMMING_LANGUAGE;
        localStorage.setItem(programmingLanguage, langSelected);
    }, [langSelected]);

    const updateFormFields = (field, value) => {
        setFormFields({ ...formFields, [field]: value });
    };

    const updateHeaders = (field, value, index) => {
        let headersList = formFields.headersList;
        headersList[index][field] = value;
        setFormFields({ ...formFields, headersList });
    };

    const addHeader = () => {
        let headersList = formFields.headersList;
        headersList.push({ key: '', value: '' });
        setFormFields({ ...formFields, headersList });
    };

    const removeHeader = (index) => {
        let headersList = formFields.headersList;
        if (headersList.length === 1) headersList[0] = { key: '', value: '' };
        else headersList.splice(index, 1);
        setFormFields({ ...formFields, headersList });
    };

    const handleAddUser = (userData) => {
        setCreateUserLoader(false);
        addUserModalFlip(false);
        updateFormFields('userName', userData.username);
    };

    const removeLineWithSubstring = (content, targetSubstring) => {
        const lines = content.split('\n');

        const filteredLines = lines.filter((line) => !line.includes(targetSubstring));

        const modifiedContent = filteredLines.join('\n');
        return modifiedContent;
    };

    const handleSelectLang = (e, isSdk = true) => {
        setLangSelected(e);
        isSdk ? changeSDKDynamicCode(e) : changeRestDynamicCode(e);
    };

    const handleSelectProtocol = (e) => {
        setProtocolSelected(e);
        if (e === 'REST') {
            changeRestDynamicCode('cURL');
            setLangSelected('cURL');
        } else {
            setTabValueRest('Generate token');
            setLangSelected('Go');
            changeSDKDynamicCode('Go');
        }
    };

    const restGWHost =
        localStorage.getItem(LOCAL_STORAGE_ENV) === 'docker'
            ? `http://localhost:${localStorage.getItem(LOCAL_STORAGE_REST_GW_PORT)}`
            : localStorage.getItem(LOCAL_STORAGE_REST_GW_HOST);

    const changeSDKDynamicCode = (lang) => {
        let codeEx = {};
        if (!SDK_CODE_EXAMPLE[lang].link) {
            codeEx.producer = SDK_CODE_EXAMPLE[lang]?.producer;
            codeEx.consumer = SDK_CODE_EXAMPLE[lang]?.consumer;
            let host =
                localStorage.getItem(LOCAL_STORAGE_ENV) === 'docker'
                    ? 'localhost'
                    : localStorage.getItem(LOCAL_STORAGE_BROKER_HOST)
                    ? localStorage.getItem(LOCAL_STORAGE_BROKER_HOST)
                    : 'memphis.memphis.svc.cluster.local';
            codeEx.producer = codeEx.producer?.replaceAll('<memphis-host>', host);
            codeEx.consumer = codeEx.consumer?.replaceAll('<memphis-host>', host);
            codeEx.producer = codeEx.producer?.replaceAll('<station-name>', stationName);
            codeEx.consumer = codeEx.consumer?.replaceAll('<station-name>', stationName);
            codeEx.producer = codeEx.producer?.replaceAll(`'<account-id>'`, parseInt(localStorage.getItem(LOCAL_STORAGE_ACCOUNT_ID)));
            codeEx.consumer = codeEx.consumer?.replaceAll(`'<account-id>'`, parseInt(localStorage.getItem(LOCAL_STORAGE_ACCOUNT_ID)));
            codeEx.producer = codeEx.producer?.replaceAll(`"<account-id>"`, parseInt(localStorage.getItem(LOCAL_STORAGE_ACCOUNT_ID)));
            codeEx.consumer = codeEx.consumer?.replaceAll(`"<account-id>"`, parseInt(localStorage.getItem(LOCAL_STORAGE_ACCOUNT_ID)));
            codeEx.producer = codeEx.producer?.replaceAll(`"<account-id>"`, parseInt(localStorage.getItem(LOCAL_STORAGE_ACCOUNT_ID)));
            codeEx.consumer = codeEx.consumer?.replaceAll(`"<account-id>"`, parseInt(localStorage.getItem(LOCAL_STORAGE_ACCOUNT_ID)));
            if (username) {
                codeEx.producer = codeEx.producer?.replaceAll('<application type username>', username);
                codeEx.consumer = codeEx.consumer?.replaceAll('<application type username>', username);
            } else if (formFields.userName !== '') {
                codeEx.producer = codeEx.producer?.replaceAll('<application type username>', formFields.userName);
                codeEx.consumer = codeEx.consumer?.replaceAll('<application type username>', formFields.userName);
            }
            if (connectionCreds) {
                codeEx.producer = codeEx.producer?.replaceAll('<broker-token>', connectionCreds);
                codeEx.consumer = codeEx.consumer?.replaceAll('<broker-token>', connectionCreds);
            } else if (formFields.password !== '') {
                codeEx.producer = codeEx.producer?.replaceAll('<broker-token>', formFields.password);
                codeEx.consumer = codeEx.consumer?.replaceAll('<broker-token>', formFields.password);
            }
            if (localStorage.getItem(LOCAL_STORAGE_USER_PASS_BASED_AUTH) === 'true') {
                codeEx.producer = codeEx.producer?.replaceAll('memphis.ConnectionToken', 'memphis.Password');
                codeEx.consumer = codeEx.consumer?.replaceAll('memphis.ConnectionToken', 'memphis.Password');
                codeEx.producer = codeEx.producer?.replaceAll('connectionToken:', 'password:');
                codeEx.consumer = codeEx.consumer?.replaceAll('connectionToken:', 'password:');
                codeEx.producer = codeEx.producer?.replaceAll('connection_token', 'password');
                codeEx.consumer = codeEx.consumer?.replaceAll('connection_token', 'password');
                codeEx.producer = codeEx.producer?.replaceAll('<broker-token>', '<password>');
                codeEx.consumer = codeEx.consumer?.replaceAll('<broker-token>', '<password>');
            } else {
                const accountId = parseInt(localStorage.getItem(LOCAL_STORAGE_ACCOUNT_ID));
                const regexPatternGo = `, memphis\.AccountId\(${accountId}\)`;
                codeEx.producer = codeEx.producer?.replaceAll(regexPatternGo, '');
                codeEx.consumer = codeEx.consumer?.replaceAll(regexPatternGo, '');
                const regexPatternJs = `accountId: ${accountId}`;
                codeEx.producer = codeEx.producer?.replaceAll(regexPatternJs, '');
                codeEx.consumer = codeEx.consumer?.replaceAll(regexPatternJs, '');
                codeEx.producer = codeEx.producer?.replaceAll(regexPatternJs, '');
                codeEx.consumer = codeEx.consumer?.replaceAll(regexPatternJs, '');
                codeEx.consumer = codeEx.consumer.replace(/^\s*[\r\n]/gm, '');
                codeEx.producer = codeEx.producer.replace(/^\s*[\r\n]/gm, '');
                const regexPatternPython = `, account_id=${accountId}`;
                codeEx.producer = codeEx.producer?.replaceAll(regexPatternPython, '');
                codeEx.consumer = codeEx.consumer?.replaceAll(regexPatternPython, '');
                const regexPatterntDotNet = `options\.AccountId = ${accountId};`;
                codeEx.producer = codeEx.producer?.replaceAll(regexPatterntDotNet, '');
                codeEx.consumer = codeEx.consumer?.replaceAll(regexPatterntDotNet, '');
            }
            if (tabValue === 'Producer' && formFields.producerConsumerName !== '') {
                codeEx.producer = codeEx.producer?.replaceAll('<producer-name>', formFields.producerConsumerName);
            }
            if (tabValue === 'Consumer' && formFields.producerConsumerName !== '') {
                codeEx.consumer = codeEx.consumer?.replaceAll('<consumer-name>', formFields.producerConsumerName);
            }
            if (tabValue === 'Consumer' && formFields.consumerGroupName !== '') {
                codeEx.consumer = codeEx.consumer?.replaceAll('<consumer-group>', formFields.consumerGroupName);
            }
            if (formFields.blocking && tabValue === 'Producer' && lang === 'Python') {
                codeEx.producer = codeEx.producer?.replaceAll('<blocking>', `, blocking=True`);
            } else codeEx.producer = codeEx.producer?.replaceAll('<blocking>', '');
            if (formFields.async && tabValue === 'Producer' && lang !== 'Python') {
                if (lang === 'Go') {
                    codeEx.producer = codeEx.producer?.replaceAll('<producer-async>', 'memphis.AsyncProduce()');
                }
                if (lang === 'TypeScript' || lang === 'Node.js') {
                    codeEx.producer = codeEx.producer?.replaceAll('<producer-async>', 'asyncProduce: true');
                }
                if (lang === '.NET (C#)') {
                    codeEx.producer = codeEx.producer?.replaceAll('<producer-async>', 'asyncProduceAck: true');
                }
            } else if (!formFields.async && tabValue === 'Producer' && lang !== 'Python') {
                if (lang === 'Go') {
                    codeEx.producer = codeEx.producer?.replaceAll('<producer-async>', 'memphis.SyncProduce()');
                }
                if (lang === 'TypeScript' || lang === 'Node.js') {
                    codeEx.producer = codeEx.producer?.replaceAll('<producer-async>', 'asyncProduce: false');
                }
                if (lang === '.NET (C#)') {
                    codeEx.producer = codeEx.producer?.replaceAll('<producer-async>', 'asyncProduceAck: false');
                }
            }
            if (formFields?.useHeaders) {
                {
                    if (lang === 'Go') {
                        codeEx.producer = codeEx.producer?.replaceAll('<headers-declaration>', 'hdrs := memphis.Headers{}');
                        codeEx.producer = codeEx.producer?.replaceAll('<headers-initiation>', 'hdrs.New()');
                        codeEx.producer = codeEx.producer?.replaceAll(
                            '<headers-addition>',
                            formFields.headersList
                                .map(
                                    (header) =>
                                        `err = hdrs.Add("${header.key === '' ? '<key>' : header.key}", "${
                                            header.value === '' ? '<value>' : header.value
                                        }")\n\tif err != nil {\n\t\tfmt.Printf("Header failed: %v", err)\n\t\tos.Exit(1)\n\t}`
                                )
                                .join('\n\t')
                        );
                    } else if (lang === 'Python') {
                        codeEx.producer = codeEx.producer?.replaceAll('<headers-initiation>', 'headers = Headers()');
                        codeEx.producer = codeEx.producer?.replaceAll(
                            '<headers-addition>',
                            formFields.headersList
                                .map((header) => `headers.add("${header.key === '' ? '<key>' : header.key}", "${header.value === '' ? '<value>' : header.value}")`)
                                .join('\n\t\t')
                        );
                    } else if (lang === '.NET (C#)') {
                        codeEx.producer = removeLineWithSubstring(codeEx.producer, '<headers-declaration>');
                        codeEx.producer = codeEx.producer?.replaceAll('<headers-initiation>', 'var commonHeaders = new NameValueCollection();');
                        codeEx.producer = codeEx.producer?.replaceAll(
                            '<headers-addition>',
                            formFields.headersList
                                .map((header) => `commonHeaders.Add("${header.key === '' ? '<key>' : header.key}", "${header.value === '' ? '<value>' : header.value}")`)
                                .join('\n\t\t\t\t')
                        );
                    } else if (lang === 'TypeScript' || lang === 'Node.js') {
                        codeEx.producer = codeEx.producer?.replaceAll('<headers-initiation>', '\n\t\tconst headers = memphis.headers()');
                        codeEx.producer = codeEx.producer?.replaceAll(
                            '<headers-addition>',
                            formFields.headersList
                                .map((header) => `headers.add("${header.key === '' ? '<key>' : header.key}", "${header.value === '' ? '<value>' : header.value}")`)
                                .join('\n\t\t\t')
                        );
                    }
                }
            } else {
                codeEx.producer = removeLineWithSubstring(codeEx.producer, '<headers-declaration>');
                codeEx.producer = removeLineWithSubstring(codeEx.producer, '<headers-initiation>');
                codeEx.producer = removeLineWithSubstring(codeEx.producer, '<headers-addition>');
                if (lang === 'Go') codeEx.producer = codeEx.producer?.replaceAll(', memphis.MsgHeaders(hdrs)', '');
                else if (lang === 'Python') codeEx.producer = codeEx.producer?.replaceAll(', headers=headers', '');
                else if (lang === '.NET (C#)') codeEx.producer = codeEx.producer?.replaceAll(', commonHeaders', '');
                else if (lang === 'TypeScript' || lang === 'Node.js') codeEx.producer = removeLineWithSubstring(codeEx.producer, 'headers: headers');
            }

            setCodeExample(codeEx);
        }
    };

    const changeRestDynamicCode = (lang) => {
        let codeEx = {};
        codeEx.producer = REST_CODE_EXAMPLE[lang].producer;
        codeEx.consumer = REST_CODE_EXAMPLE[lang].consumer;
        codeEx.tokenGenerate = REST_CODE_EXAMPLE[lang].tokenGenerate;
        codeEx.producer = codeEx.producer.replaceAll('localhost', restGWHost);
        codeEx.producer = codeEx.producer.replaceAll('<station-name>', stationName);
        codeEx.consumer = codeEx.consumer.replaceAll('localhost', restGWHost);
        codeEx.consumer = codeEx.consumer.replaceAll('<station-name>', stationName);
        codeEx.tokenGenerate = codeEx.tokenGenerate.replaceAll('localhost', restGWHost);
        codeEx.tokenGenerate = codeEx.tokenGenerate.replaceAll(`"<account-id>"`, parseInt(localStorage.getItem(LOCAL_STORAGE_ACCOUNT_ID)));
        if (username) {
            codeEx.tokenGenerate = codeEx.tokenGenerate?.replaceAll('<application type username>', username);
        } else if (formFields.userName !== '') {
            codeEx.tokenGenerate = codeEx.tokenGenerate?.replaceAll('<application type username>', formFields.userName);
        }
        if (connectionCreds) {
            codeEx.tokenGenerate = codeEx.tokenGenerate?.replaceAll('<broker-token>', connectionCreds);
        }
        if (localStorage.getItem(LOCAL_STORAGE_USER_PASS_BASED_AUTH) === 'true') {
            codeEx.tokenGenerate = codeEx.tokenGenerate?.replaceAll('connection_token', 'password');
            codeEx.tokenGenerate = codeEx.tokenGenerate?.replaceAll('<broker-token>', '<password>');
            codeEx.tokenGenerate = codeEx.tokenGenerate?.replaceAll('memphis.ConnectionToken', 'memphis.Password');
            codeEx.tokenGenerate = codeEx.tokenGenerate?.replaceAll("strings.NewReader('{", 'strings.NewReader(`{');
            codeEx.tokenGenerate = codeEx.tokenGenerate?.replaceAll("}')", ' }`)');
        }
        if (formFields.password !== '') {
            codeEx.tokenGenerate = codeEx.tokenGenerate?.replaceAll('<password>', formFields.password);
        }
        if (formFields.producerConsumerName !== '') {
            codeEx.consumer = codeEx.consumer?.replaceAll('<consumer-name>', formFields.producerConsumerName);
        }
        if (formFields.consumerGroupName !== '') {
            codeEx.consumer = codeEx.consumer?.replaceAll('<consumer-group>', formFields.consumerGroupName);
        }
        if (formFields.tokenExpiry !== '') {
            codeEx.tokenGenerate = codeEx.tokenGenerate?.replaceAll('<token_expiry_in_minutes>', formFields.tokenExpiry);
        }
        if (formFields.refreshToken !== '') {
            codeEx.tokenGenerate = codeEx.tokenGenerate?.replaceAll('<refresh_token_expiry_in_minutes>', formFields.refreshToken);
        }
        if (formFields.tokenExpiry === '') {
            codeEx.tokenGenerate = codeEx.tokenGenerate?.replaceAll('<token_expiry_in_minutes>', 100);
        }
        if (formFields.refreshToken === '') {
            codeEx.tokenGenerate = codeEx.tokenGenerate?.replaceAll('<refresh_token_expiry_in_minutes>', 100);
        }
        if (formFields.batchSize !== '') {
            codeEx.consumer = codeEx.consumer?.replaceAll('<batch-size>', formFields.batchSize);
        }
        if (formFields.batchTimeout !== '') {
            codeEx.consumer = codeEx.consumer?.replaceAll('<batch-max-wait-time-ms>', formFields.batchTimeout);
        }
        if (formFields.batchSize === '') {
            codeEx.consumer = codeEx.consumer?.replaceAll('<batch-size>', 10);
        }
        if (formFields.batchTimeout === '') {
            codeEx.tokenGenerate = codeEx.tokenGenerate?.replaceAll('<batch-max-wait-time-ms>', 1000);
        }
        if (formFields.jwt !== '') {
            codeEx.producer = codeEx.producer?.replaceAll('<jwt>', formFields.jwt);
            codeEx.consumer = codeEx.consumer?.replaceAll('<jwt>', formFields.jwt);
        }

        if (formFields.useHeaders) {
            if (lang === 'cURL') {
                codeEx.producer = codeEx.producer?.replaceAll(
                    '<headers-addition>',
                    formFields.headersList
                        .map((header) => `--header '${header.key === '' ? '<key>' : header.key}: ${header.value === '' ? '<value>' : header.value}' \\`)
                        .join(' \n')
                );
            } else if (lang === 'Node.js') {
                codeEx.producer = codeEx.producer?.replaceAll(
                    '<headers-addition>',
                    formFields.headersList
                        .map((header) => `'${header.key === '' ? '<key>' : header.key}': '${header.value === '' ? '<value>' : header.value}',`)
                        .join('\n\t')
                );
            } else if (lang === 'Go') {
                codeEx.producer = codeEx.producer?.replaceAll(
                    '<headers-addition>',
                    formFields.headersList
                        .map((header) => `req.Header.Add("${header.key === '' ? '<key>' : header.key}", "${header.value === '' ? '<value>' : header.value}")`)
                        .join('\n\t\t')
                );
            } else if (lang === 'Java') {
                codeEx.producer = codeEx.producer?.replaceAll(
                    '<headers-addition>',
                    formFields.headersList
                        .map((header) => `.addHeader("${header.key === '' ? '<key>' : header.key}", "${header.value === '' ? '<value>' : header.value}")`)
                        .join('\n  ')
                );
            } else if (lang === 'Python') {
                codeEx.producer = codeEx.producer?.replaceAll(
                    '<headers-addition>',
                    formFields.headersList
                        .map((header) => `'${header.key === '' ? '<key>' : header.key}': '${header.value === '' ? '<value>' : header.value}',`)
                        .join('\n  ')
                );
            } else if (lang === 'JavaScript - jQuery') {
                codeEx.producer = codeEx.producer?.replaceAll(
                    '<headers-addition>',
                    formFields.headersList
                        .map((header) => `"${header.key === '' ? '<key>' : header.key}": "${header.value === '' ? '<value>' : header.value}",`)
                        .join('\n\t')
                );
            } else if (lang === 'JavaScript - Fetch') {
                codeEx.producer = codeEx.producer?.replaceAll(
                    '<headers-addition>',
                    formFields.headersList
                        .map((header) => `myHeaders.append("${header.key === '' ? '<key>' : header.key}", "${header.value === '' ? '<value>' : header.value}");`)
                        .join('\n')
                );
            }
        } else {
            codeEx.producer = removeLineWithSubstring(codeEx.producer, '<headers-addition>');
        }
        setCodeExample(codeEx);
    };

    const generateEditor = (langCode, value) => {
        return (
            <>
                <Editor
                    options={{
                        minimap: { enabled: false },
                        scrollbar: { verticalScrollbarSize: 0, horizontalScrollbarSize: 0 },
                        scrollBeyondLastLine: false,
                        roundedSelection: false,
                        formatOnPaste: true,
                        formatOnType: true,
                        readOnly: true,
                        fontSize: '12px',
                        fontFamily: 'Inter'
                    }}
                    language={langCode}
                    height="calc(100% - 10px)"
                    width="calc(100% - 25px)"
                    value={value}
                />
                <Copy data={value} />
            </>
        );
    };

    return (
        <div className="code-example-details-container sdk-example">
            <div className="left-side-container">
                {withHeader && (
                    <div className="modal-header">
                        <div className="header-img-container">
                            <CodeIcon className="headerImage" alt="codeIcon" />
                        </div>
                        <p className="modal-title-sdk">Client generator</p>
                        <label>Utilize the client generator for a quick integration of your application with the station.</label>
                    </div>
                )}
                <div className="code-generator-container" style={{ height: withHeader ? 'calc(100% - 150px)' : '700px' }}>
                    <div className="username-section">
                        <span className="input-item">
                            <p className="field-title">Communication type</p>
                            <SelectComponent
                                value={protocolSelected}
                                fontSize="14px"
                                colorType="navy"
                                backgroundColorType="none"
                                borderColorType="gray"
                                radiusType="semi-round"
                                height="42px"
                                options={selectProtocolOption}
                                onChange={(e) => handleSelectProtocol(e)}
                                popupClassName="select-options"
                            />
                        </span>
                        <span className="input-item">
                            <p className="field-title">Language</p>
                            <SelectComponent
                                value={langSelected}
                                colorType="navy"
                                fontSize="14px"
                                backgroundColorType="none"
                                borderColorType="gray"
                                radiusType="semi-round"
                                height="42px"
                                options={protocolSelected === 'SDK' ? sdkLangOptions : restLangOptions}
                                onChange={(e) => (protocolSelected === 'SDK' ? handleSelectLang(e) : handleSelectLang(e, false))}
                                popupClassName="select-options"
                            />
                        </span>
                    </div>

                    <>
                        {protocolSelected === 'SDK' && (
                            <>
                                <div className="installation">
                                    <p className="field-title">Package installation</p>
                                    <div className="install-copy">
                                        <p>{SDK_CODE_EXAMPLE[langSelected].installation}</p>
                                        <Copy data={SDK_CODE_EXAMPLE[langSelected].installation}/>
                                    </div>
                                </div>

                                <div className="ant-divider" style={{marginBottom: '.5rem', marginTop: '1.5rem'}} role="separator"></div>
                            </>
                        )}
                        {protocolSelected === 'REST' && (
                            <SegmentButton value={tabValueRest} options={tabsProtocol} onChange={(tabValueRest) => setTabValueRest(tabValueRest)} size="medium" />
                        )}
                        {
                            <div className="code-builder">
                                <div className="parameters-section">
                                    {(tabValue === 'SDK' || tabValueRest === 'Generate token') && withHeader && (
                                        <div className="username-section">
                                            <span className="input-item">
                                                <TitleComponent headerTitle="Username" typeTitle="sub-header"/>
                                                <Form.Item>
                                                    <CustomSelect
                                                        placeholder={'Select user'}
                                                        value={formFields.userName}
                                                        options={users}
                                                        onChange={(e) => updateFormFields('userName', e)}
                                                        type="user"
                                                        handleCreateNew={() => addUserModalFlip(true)}
                                                        showCreatedBy={false}
                                                    />
                                                </Form.Item>
                                            </span>
                                            <span className="input-item">
                                                <TitleComponent headerTitle="Password" typeTitle="sub-header"/>
                                                <Form.Item name="password">
                                                    <Input
                                                        placeholder="Type password"
                                                        type="password"
                                                        fontSize="14px"
                                                        radiusType="semi-round"
                                                        colorType="black"
                                                        backgroundColorType="none"
                                                        borderColorType="gray"
                                                        height="40px"
                                                        onBlur={(e) => updateFormFields('password', e.target.value)}
                                                        onChange={(e) => updateFormFields('password', e.target.value)}
                                                        value={formFields.password}
                                                    />
                                                </Form.Item>
                                            </span>
                                        </div>
                                    )}

                                    {protocolSelected === 'REST' && tabValueRest === 'Generate token' && (
                                        <>
                                            <div className="username-section">
                                                        <span className="input-item">
                                                            <TitleComponent headerTitle="Token expiry"
                                                                            spanHeader="(In minutes)"
                                                                            typeTitle="sub-header"/>
                                                            <Form.Item>
                                                                <Input
                                                                    placeholder="Type token expiry"
                                                                    type="text"
                                                                    fontSize="14px"
                                                                    maxLength="128"
                                                                    radiusType="semi-round"
                                                                    colorType="black"
                                                                    backgroundColorType="white"
                                                                    borderColorType="gray"
                                                                    height="40px"
                                                                    onBlur={(e) => !isNaN(e.target.value) && updateFormFields('tokenExpiry', e.target.value)}
                                                                    onChange={(e) => !isNaN(e.target.value) && updateFormFields('tokenExpiry', e.target.value)}
                                                                    value={formFields.tokenExpiry}
                                                                />
                                                            </Form.Item>
                                                        </span>
                                                <span className="input-item">
                                                            <TitleComponent headerTitle="Refresh token expiry"
                                                                            spanHeader="(In minutes)"
                                                                            typeTitle="sub-header"/>
                                                            <Form.Item>
                                                                <Input
                                                                    placeholder="Refresh token expiry"
                                                                    type="text"
                                                                    fontSize="14px"
                                                                    maxLength="128"
                                                                    radiusType="semi-round"
                                                                    colorType="black"
                                                                    backgroundColorType="white"
                                                                    borderColorType="gray"
                                                                    height="40px"
                                                                    onBlur={(e) => !isNaN(e.target.value) && updateFormFields('refreshToken', e.target.value)}
                                                                    onChange={(e) => !isNaN(e.target.value) && updateFormFields('refreshToken', e.target.value)}
                                                                    value={formFields.refreshToken}
                                                                />
                                                            </Form.Item>
                                                        </span>
                                            </div>
                                        </>
                                    )}
                                    {protocolSelected === 'REST' && tabValueRest !== 'Generate token' && (
                                        <>
                                            <TitleComponent
                                                headerTitle="JWT"
                                                typeTitle="sub-header"
                                                headerDescription="To be able to recognize a specific producer across the system"
                                            />
                                            <Form.Item>
                                                <Input
                                                    placeholder="JWT"
                                                    type="text"
                                                    fontSize="14px"
                                                    radiusType="semi-round"
                                                    colorType="black"
                                                    backgroundColorType="white"
                                                    borderColorType="gray"
                                                    height="40px"
                                                    onBlur={(e) => updateFormFields('jwt', e.target.value)}
                                                    onChange={(e) => updateFormFields('jwt', e.target.value)}
                                                    value={formFields.jwt}
                                                />
                                            </Form.Item>
                                        </>
                                    )}
                                    {(protocolSelected === 'SDK' || tabValueRest === 'Consume data') && (
                                        <>
                                            <div className="username-section">
                                                        <span className="input-item">
                                                            <TitleComponent
                                                                headerTitle={`${
                                                                    tabValue === 'Producer' && tabValueRest !== 'Consume data' ? 'Producer' : 'Consumer'
                                                                } name`}
                                                                typeTitle="sub-header"
                                                            />
                                                            <Form.Item>
                                                                <Input
                                                                    placeholder={`Type ${
                                                                        tabValue === 'Producer' && tabValueRest !== 'Consume data' ? 'producer' : 'consumer'
                                                                    } name`}
                                                                    type="text"
                                                                    fontSize="14px"
                                                                    maxLength="128"
                                                                    radiusType="semi-round"
                                                                    colorType="black"
                                                                    backgroundColorType="white"
                                                                    borderColorType="gray"
                                                                    height="40px"
                                                                    onBlur={(e) => updateFormFields('producerConsumerName', e.target.value)}
                                                                    onChange={(e) => updateFormFields('producerConsumerName', e.target.value)}
                                                                    value={formFields.producerConsumerName}
                                                                />
                                                            </Form.Item>
                                                        </span>
                                                {(tabValue === 'Consumer' || tabValueRest === 'Consume data') && (
                                                    <span className="input-item">
                                                                <TitleComponent headerTitle="Consumer group name"
                                                                                typeTitle="sub-header"/>
                                                                <Form.Item>
                                                                    <Input
                                                                        placeholder="Type consumer group name"
                                                                        type="text"
                                                                        fontSize="14px"
                                                                        maxLength="128"
                                                                        radiusType="semi-round"
                                                                        colorType="black"
                                                                        backgroundColorType="white"
                                                                        borderColorType="gray"
                                                                        height="40px"
                                                                        onBlur={(e) => updateFormFields('consumerGroupName', e.target.value)}
                                                                        onChange={(e) => updateFormFields('consumerGroupName', e.target.value)}
                                                                        value={formFields.consumerGroupName}
                                                                    />
                                                                </Form.Item>
                                                            </span>
                                                )}
                                            </div>
                                        </>
                                    )}
                                    {protocolSelected === 'REST' && tabValueRest === 'Consume data' && (
                                        <>
                                            <div className="username-section">
                                                        <span className="input-item">
                                                            <TitleComponent headerTitle="Batch size"
                                                                            typeTitle="sub-header"/>
                                                            <Form.Item>
                                                                <Input
                                                                    placeholder="Type batch size"
                                                                    type="number"
                                                                    fontSize="14px"
                                                                    maxLength="128"
                                                                    radiusType="semi-round"
                                                                    colorType="black"
                                                                    backgroundColorType="white"
                                                                    borderColorType="gray"
                                                                    height="40px"
                                                                    onBlur={(e) => updateFormFields('batchSize', e.target.value)}
                                                                    onChange={(e) => updateFormFields('batchSize', e.target.value)}
                                                                    value={formFields.batchSize}
                                                                />
                                                            </Form.Item>
                                                        </span>
                                                <span className="input-item">
                                                            <TitleComponent headerTitle="Batch timeout"
                                                                            spanHeader="(ms)" typeTitle="sub-header"/>
                                                            <Form.Item>
                                                                <Input
                                                                    placeholder="Batch timeout (ms)"
                                                                    type="text"
                                                                    fontSize="14px"
                                                                    maxLength="128"
                                                                    radiusType="semi-round"
                                                                    colorType="black"
                                                                    backgroundColorType="white"
                                                                    borderColorType="gray"
                                                                    height="40px"
                                                                    onBlur={(e) => updateFormFields('batchTimeout', e.target.value)}
                                                                    onChange={(e) => updateFormFields('batchTimeout', e.target.value)}
                                                                    value={formFields.batchTimeout}
                                                                />
                                                            </Form.Item>
                                                        </span>
                                            </div>
                                        </>
                                    )}
                                    {protocolSelected === 'SDK' && langSelected === 'Python' && tabValue === 'Producer' && (
                                        <div className="username-section">
                                            <TitleComponent
                                                headerTitle="Blocking"
                                                typeTitle="sub-header"
                                                headerDescription="For better performance, the client won't block requests while waiting for an acknowledgment"
                                            />

                                            <Form.Item>
                                                <Switcher
                                                    onChange={() => updateFormFields('blocking', !formFields.blocking)}
                                                    checked={formFields.blocking}/>
                                            </Form.Item>
                                        </div>
                                    )}
                                    {protocolSelected === 'SDK' && langSelected !== 'Python' && tabValue === 'Producer' && (
                                        <div className="username-section">
                                            <TitleComponent
                                                headerTitle="Async"
                                                typeTitle="sub-header"
                                                headerDescription="For better performance, the client won't block requests while waiting for an acknowledgment"
                                            />

                                            <Form.Item>
                                                <Switcher onChange={() => updateFormFields('async', !formFields.async)}
                                                          checked={formFields.async}/>
                                            </Form.Item>
                                        </div>
                                    )}
                                    {((protocolSelected === 'SDK' && tabValue === 'Producer') ||
                                        (protocolSelected === 'REST' && tabValueRest === 'Produce data')) && (
                                        <div className="username-section">
                                            <TitleComponent headerTitle="Headers" typeTitle="sub-header"
                                                            headerDescription="Add header to the message"/>
                                            <Form.Item>
                                                <Switcher
                                                    onChange={() => updateFormFields('useHeaders', !formFields.useHeaders)}
                                                    checked={formFields.useHeaders}
                                                />
                                            </Form.Item>
                                        </div>
                                    )}
                                    {formFields.useHeaders &&
                                        ((protocolSelected === 'SDK' && tabValue === 'Producer') ||
                                            (protocolSelected === 'REST' && tabValueRest === 'Produce data')) && (
                                            <div>
                                                {formFields.headersList.map((header, index) => (
                                                    <div className="header-section" key={index}>
                                                        <span className="input-item">
                                                            <Form.Item>
                                                                <Input
                                                                    placeholder="key"
                                                                    type="text"
                                                                    fontSize="14px"
                                                                    maxLength="200"
                                                                    radiusType="semi-round"
                                                                    colorType="black"
                                                                    backgroundColorType="white"
                                                                    borderColorType="gray"
                                                                    height="40px"
                                                                    onBlur={(e) => updateHeaders('key', e.target.value, index)}
                                                                    onChange={(e) => updateHeaders('key', e.target.value, index)}
                                                                    value={header.key}
                                                                />
                                                            </Form.Item>
                                                        </span>
                                                        <span className="input-item">
                                                            <Form.Item>
                                                                <Input
                                                                    placeholder="value"
                                                                    type="text"
                                                                    maxLength="200"
                                                                    fontSize="14px"
                                                                    radiusType="semi-round"
                                                                    colorType="black"
                                                                    backgroundColorType="white"
                                                                    borderColorType="gray"
                                                                    height="40px"
                                                                    onBlur={(e) => updateHeaders('value', e.target.value, index)}
                                                                    onChange={(e) => updateHeaders('value', e.target.value, index)}
                                                                    value={header.value}
                                                                />
                                                            </Form.Item>
                                                        </span>
                                                        <FiMinusCircle className="remove-icon"
                                                                       onClick={() => removeHeader(index)}/>
                                                    </div>
                                                ))}
                                                <div className="generate-action" onClick={() => addHeader()}>
                                                    <FiPlus/>
                                                    <span>Add more</span>
                                                </div>
                                            </div>
                                        )}
                                </div>
                            </div>
                        }
                    </>
                </div>
                <Modal
                    header="Generate JWT token"
                    displayButtons={false}
                    width="460px"
                    clickOutside={() => setGenerateModal(false)}
                    open={generateModal}
                    className="generate-modal"
                >
                    <GenerateTokenModal
                        host={restGWHost}
                        close={() => {
                            setGenerateModal(false);
                            setTabValueRest('Produce data');
                        }}
                        returnToken={(e) => updateFormFields('jwt', e.jwt)}
                    />
                </Modal>

                <Drawer
                    placement="right"
                    title="Add a new user"
                    onClose={() => {
                        setCreateUserLoader(false);
                        addUserModalFlip(false);
                    }}
                    destroyOnClose={true}
                    width="650px"
                    open={addUserModalIsOpen}
                >
                    <CreateUserDetails
                        clientType
                        createUserRef={createUserRef}
                        closeModal={(userData) => handleAddUser(userData)}
                        handleLoader={(e) => setCreateUserLoader(e)}
                        isLoading={createUserLoader}
                    />
                </Drawer>
            </div>
            <Divider type="vertical"/>
            <div>
                <div className={`code-output-title ${withHeader && 'code-output-title-code-example'}`}>
                    <p>Generated code</p>
                </div>
                {protocolSelected === 'SDK' && SDK_CODE_EXAMPLE[langSelected]?.link && (
                    <div className="guidline">
                        <NoCodeExampleIcon/>
                        <div className="content">
                            <p>{SDK_CODE_EXAMPLE[langSelected].title}</p>
                            <span>{SDK_CODE_EXAMPLE[langSelected].desc}</span>
                            <a className="learn-more" href={SDK_CODE_EXAMPLE[langSelected].link} target="_blank">
                                View Documentation
                            </a>
                        </div>
                    </div>
                )}
                {protocolSelected === 'SDK' && !SDK_CODE_EXAMPLE[langSelected]?.link && (
                    <>
                        <div className="tabs">
                            <div className="code-example">
                                <div className="code-content">
                                    {generateEditor(SDK_CODE_EXAMPLE[langSelected].langCode, tabValue === 'Consumer' ? codeExample.consumer : codeExample.producer)}
                                </div>
                            </div>
                        </div>
                    </>
                )}
                {protocolSelected === 'REST' && (
                    <>
                        {tabValueRest === 'Generate token' && (
                            <div className="installation">
                                <div className="generate-wrapper">
                                    <p className="field-title">Generate a token</p>
                                    <div className="generate-action" onClick={() => setGenerateModal(true)}>
                                        <RefreshIcon width="14"/>
                                        <span>Generate JWT token</span>
                                    </div>
                                </div>
                                <div className="code-example ce-protoco">
                                    <div
                                        className="code-content">{generateEditor(REST_CODE_EXAMPLE[langSelected].langCode, codeExample.tokenGenerate)}</div>
                                </div>
                            </div>
                        )}
                        {tabValueRest === 'Produce data' && (
                            <div className="tabs">
                                <p className="field-title">Produce data</p>
                                <div className="code-example ce-protoco">
                                    <div
                                        className="code-content produce">{generateEditor(REST_CODE_EXAMPLE[langSelected].langCode, codeExample.producer)}</div>
                                </div>
                            </div>
                        )}
                        {tabValueRest === 'Consume data' && (
                            <div className="tabs">
                                <p className="field-title">Consumem data</p>
                                <div className="code-example ce-protoco">
                                    <div
                                        className="code-content produce">{generateEditor(REST_CODE_EXAMPLE[langSelected].langCode, codeExample.consumer)}</div>
                                </div>
                            </div>
                        )}
                    </>
                )}
            </div>
        </div>
    );
};

export default SdkExample;
