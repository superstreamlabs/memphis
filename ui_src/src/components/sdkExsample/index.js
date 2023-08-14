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

import Editor, { loader } from '@monaco-editor/react';
import React, { useEffect, useState } from 'react';
import * as monaco from 'monaco-editor';

import { PROTOCOL_CODE_EXAMPLE, SDK_CODE_EXAMPLE, selectLngOption, selectProtocolLngOptions } from '../../const/codeExample';
import {
    LOCAL_STORAGE_ACCOUNT_ID,
    LOCAL_STORAGE_BROKER_HOST,
    LOCAL_STORAGE_ENV,
    LOCAL_STORAGE_REST_GW_HOST,
    LOCAL_STORAGE_REST_GW_PORT,
    LOCAL_STORAGE_USER_PASS_BASED_AUTH
} from '../../const/localStorageConsts';
import GenerateTokenModal from '../../domain/stationOverview/components/generateTokenModal';
import noCodeExample from '../../assets/images/noCodeExample.svg';
import codeIcon from '../../assets/images/codeIcon.svg';
import refresh from '../../assets/images/refresh.svg';
import SelectComponent from '../select';
import CustomTabs from '../Tabs';
import Modal from '../modal';
import Copy from '../copy';

loader.init();
loader.config({ monaco });

const tabs = ['Producer', 'Consumer'];
const selectProtocolOption = ['SDK (TCP)', 'REST (HTTP)'];

const SdkExample = ({ consumer, showTabs = true, stationName, username, connectionCreds, withHeader = false }) => {
    const [langSelected, setLangSelected] = useState('Go');
    const [protocolSelected, setProtocolSelected] = useState('SDK (TCP)');
    const [codeExample, setCodeExample] = useState({
        producer: '',
        consumer: ''
    });
    const [tabValue, setTabValue] = useState(consumer ? 'Consumer' : 'Producer');
    const [generateModal, setGenerateModal] = useState(false);

    const restGWHost =
        localStorage.getItem(LOCAL_STORAGE_ENV) === 'docker'
            ? `http://localhost:${localStorage.getItem(LOCAL_STORAGE_REST_GW_PORT)}`
            : localStorage.getItem(LOCAL_STORAGE_REST_GW_HOST);

    const changeDynamicCode = (lang) => {
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
            }
            if (connectionCreds) {
                codeEx.producer = codeEx.producer?.replaceAll('<broker-token>', connectionCreds);
                codeEx.consumer = codeEx.consumer?.replaceAll('<broker-token>', connectionCreds);
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
            setCodeExample(codeEx);
        }
    };

    const changeProtocolDynamicCode = (lang) => {
        let codeEx = {};
        codeEx.producer = PROTOCOL_CODE_EXAMPLE[lang].producer;
        codeEx.tokenGenerate = PROTOCOL_CODE_EXAMPLE[lang].tokenGenerate;
        codeEx.producer = codeEx.producer.replaceAll('localhost', restGWHost);
        codeEx.producer = codeEx.producer.replaceAll('<station-name>', stationName);
        codeEx.tokenGenerate = codeEx.tokenGenerate.replaceAll('localhost', restGWHost);
        codeEx.producer = codeEx.producer.replaceAll(`"<account-id>"`, parseInt(localStorage.getItem(LOCAL_STORAGE_ACCOUNT_ID)));
        codeEx.tokenGenerate = codeEx.tokenGenerate.replaceAll(`"<account-id>"`, parseInt(localStorage.getItem(LOCAL_STORAGE_ACCOUNT_ID)));
        if (username) {
            codeEx.tokenGenerate = codeEx.tokenGenerate?.replaceAll('<application type username>', username);
        }
        if (connectionCreds) {
            codeEx.tokenGenerate = codeEx.tokenGenerate?.replaceAll('<broker-token>', connectionCreds);
        }
        if (localStorage.getItem(LOCAL_STORAGE_USER_PASS_BASED_AUTH) === 'true') {
            codeEx.tokenGenerate = codeEx.tokenGenerate?.replaceAll('connection_token', 'password');
            codeEx.tokenGenerate = codeEx.tokenGenerate?.replaceAll('<broker-token>', '<password>');
            codeEx.tokenGenerate = codeEx.tokenGenerate?.replaceAll('memphis.ConnectionToken', 'memphis.Password');
        }
        setCodeExample(codeEx);
    };

    useEffect(() => {
        protocolSelected === 'SDK (TCP)' ? changeDynamicCode(langSelected) : changeProtocolDynamicCode('cURL');
    }, []);

    const handleSelectLang = (e, isSdk = true) => {
        setLangSelected(e);
        isSdk ? changeDynamicCode(e) : changeProtocolDynamicCode(e);
    };

    const handleSelectProtocol = (e) => {
        setProtocolSelected(e);
        if (e === 'REST (HTTP)') {
            changeProtocolDynamicCode('cURL');
            setLangSelected('cURL');
        } else {
            setLangSelected('Go');
            changeDynamicCode('Go');
        }
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
            {withHeader && (
                <div className="modal-header">
                    <div className="header-img-container">
                        <img className="headerImage" src={codeIcon} alt="codeIcon" />
                    </div>
                    <p>Code examplesn</p>
                    <label>Some code snippets that will help you get started with Memphis</label>
                </div>
            )}
            <div className="select-lan">
                <div>
                    <p className="field-title">Protocol</p>
                    <SelectComponent
                        value={protocolSelected}
                        colorType="navy"
                        backgroundColorType="none"
                        borderColorType="gray"
                        radiusType="semi-round"
                        width="220px"
                        height="50px"
                        options={selectProtocolOption}
                        onChange={(e) => handleSelectProtocol(e)}
                        popupClassName="select-options"
                    />
                </div>
                <div>
                    <p className="field-title">Language</p>
                    <SelectComponent
                        value={langSelected}
                        colorType="navy"
                        backgroundColorType="none"
                        borderColorType="gray"
                        radiusType="semi-round"
                        width="220px"
                        height="50px"
                        options={protocolSelected === 'SDK (TCP)' ? selectLngOption : selectProtocolLngOptions}
                        onChange={(e) => (protocolSelected === 'SDK (TCP)' ? handleSelectLang(e) : handleSelectLang(e, false))}
                        popupClassName="select-options"
                    />
                </div>
            </div>
            {protocolSelected === 'SDK (TCP)' && (
                <>
                    <div className="installation">
                        <p className="field-title">Package installation</p>
                        <div className="install-copy">
                            <p>{SDK_CODE_EXAMPLE[langSelected].installation}</p>
                            <Copy data={SDK_CODE_EXAMPLE[langSelected].installation} />
                        </div>
                    </div>
                    <div className="tabs">
                        {showTabs && <CustomTabs value={tabValue} onChange={(tabValue) => setTabValue(tabValue)} tabs={tabs}></CustomTabs>}
                        {!showTabs && <p className="field-title">{`Code snippet for ${tabValue === 'Producer' ? 'producing' : 'consuming'} data`}</p>}
                        {SDK_CODE_EXAMPLE[langSelected].link ? (
                            <div className="guidline">
                                <img src={noCodeExample} />
                                <div className="content">
                                    <p>{SDK_CODE_EXAMPLE[langSelected].title}</p>
                                    <span>{SDK_CODE_EXAMPLE[langSelected].desc}</span>
                                    <a className="learn-more" href={SDK_CODE_EXAMPLE[langSelected].link} target="_blank">
                                        View Documentation
                                    </a>
                                </div>
                            </div>
                        ) : (
                            <div className="code-example">
                                <div className="code-content">
                                    {generateEditor(SDK_CODE_EXAMPLE[langSelected].langCode, tabValue === 'Consumer' ? codeExample.consumer : codeExample.producer)}
                                </div>
                            </div>
                        )}
                    </div>
                </>
            )}
            {protocolSelected === 'REST (HTTP)' && (
                <>
                    <div className="installation">
                        <div className="generate-wrapper">
                            <p className="field-title">Step 1: Generate a token</p>
                            <div className="generate-action" onClick={() => setGenerateModal(true)}>
                                <img src={refresh} width="14" />
                                <span>Generate JWT token</span>
                            </div>
                        </div>
                        <div className="code-example ce-protoco">
                            <div className="code-content">{generateEditor(PROTOCOL_CODE_EXAMPLE[langSelected].langCode, codeExample.tokenGenerate)}</div>
                        </div>
                    </div>
                    <div className="tabs">
                        <p className="field-title">{`Step 2: ${consumer ? 'Consume' : 'Produce'} data`}</p>
                        {consumer ? (
                            <div className="guidline">
                                <img src={noCodeExample} />
                                <div className="content">
                                    <p>Coming soon</p>
                                    <span>
                                        Please <a>upvote</a> to prioritize it!
                                    </span>
                                </div>
                            </div>
                        ) : (
                            <div className="code-example ce-protoco">
                                <div className="code-content produce">{generateEditor(PROTOCOL_CODE_EXAMPLE[langSelected].langCode, codeExample.producer)}</div>
                            </div>
                        )}
                    </div>
                </>
            )}
            <Modal
                header="Generate JWT token"
                displayButtons={false}
                height="400px"
                width="400px"
                clickOutside={() => setGenerateModal(false)}
                open={generateModal}
                className="generate-modal"
            >
                <GenerateTokenModal host={restGWHost} close={() => setGenerateModal(false)} />
            </Modal>
        </div>
    );
};

export default SdkExample;
