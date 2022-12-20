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

import './style.scss';

import React, { useEffect, useState } from 'react';
import { loader } from '@monaco-editor/react';
import Editor from '@monaco-editor/react';
import * as monaco from 'monaco-editor';

import { LOCAL_STORAGE_ENV, LOCAL_STORAGE_NAMESPACE } from '../../../../const/localStorageConsts';
import { PROTOCOL_CODE_EXAMPLE, SDK_CODE_EXAMPLE } from '../../../../const/codeExample';
import SelectComponent from '../../../../components/select';
import CustomTabs from '../../../../components/Tabs';
import Copy from '../../../../components/copy';

loader.config({ monaco });

const tabs = ['Producer', 'Consumer'];

const SdkExample = ({ consumer, showTabs = true }) => {
    const [langSelected, setLangSelected] = useState('Go');
    const [protocolSelected, setProtocolSelected] = useState('SDK(TCP)');
    const selectLngOption = ['Go', 'Node.js', 'Typescript', 'Python'];
    const selectProtocolOption = ['SDK(TCP)', 'Rest'];
    const [codeExample, setCodeExample] = useState({
        import: '',
        connect: '',
        producer: '',
        consumer: ''
    });
    const [tabValue, setTabValue] = useState(consumer ? 'Consumer' : 'Producer');

    const url = window.location.href;
    const stationName = url.split('stations/')[1];

    const changeDynamicCode = (lang) => {
        let codeEx = {};
        codeEx.producer = SDK_CODE_EXAMPLE[lang].producer;
        codeEx.consumer = SDK_CODE_EXAMPLE[lang].consumer;
        let host = process.env.REACT_APP_SANDBOX_ENV
            ? 'broker.sandbox.memphis.dev'
            : localStorage.getItem(LOCAL_STORAGE_ENV) === 'docker'
            ? 'localhost'
            : 'memphis-cluster.' + localStorage.getItem(LOCAL_STORAGE_NAMESPACE) + '.svc.cluster.local';
        codeEx.producer = codeEx.producer.replaceAll('<memphis-host>', host);
        codeEx.consumer = codeEx.consumer.replaceAll('<memphis-host>', host);
        codeEx.producer = codeEx.producer.replaceAll('<station-name>', stationName);
        codeEx.consumer = codeEx.consumer.replaceAll('<station-name>', stationName);
        setCodeExample(codeEx);
    };

    const changeProtocolDynamicCode = (lang) => {
        let codeEx = {};
        codeEx.producer = PROTOCOL_CODE_EXAMPLE[lang].producer;
        codeEx.tokenGenerate = PROTOCOL_CODE_EXAMPLE[lang].tokenGenerate;
        let host =
            localStorage.getItem(LOCAL_STORAGE_ENV) === 'docker'
                ? 'localhost'
                : 'memphis-http-proxy.' + localStorage.getItem(LOCAL_STORAGE_NAMESPACE) + '.svc.cluster.local';
        codeEx.producer = codeEx.producer.replaceAll('localhost', host);
        codeEx.tokenGenerate = codeEx.tokenGenerate.replaceAll('localhost', host);

        setCodeExample(codeEx);
    };

    useEffect(() => {
        changeDynamicCode(langSelected);
    }, []);

    const handleSelectLang = (e) => {
        setLangSelected(e);
        changeDynamicCode(e);
    };
    const handleSelectProtocol = (e) => {
        setProtocolSelected(e);
        changeProtocolDynamicCode(e);
    };

    return (
        <div className="code-example-details-container sdk-example">
            <div className="header-wrapper">
                <p>Code example</p>
                <span>We'll provide you with snippets that you can easily connect your application with Memphis</span>
            </div>
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
                        options={selectLngOption}
                        onChange={(e) => handleSelectLang(e)}
                        popupClassName="select-options"
                    />
                </div>
            </div>
            {protocolSelected === 'SDK(TCP)' && (
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
                        {tabValue === 'Producer' && (
                            <div className="code-example">
                                <div className="code-content">
                                    <Editor
                                        options={{
                                            minimap: { enabled: false },
                                            scrollbar: { verticalScrollbarSize: 0 },
                                            scrollBeyondLastLine: false,
                                            roundedSelection: false,
                                            formatOnPaste: true,
                                            formatOnType: true,
                                            readOnly: true,
                                            fontSize: '14px',
                                            fontFamily: 'Inter'
                                        }}
                                        language={SDK_CODE_EXAMPLE[langSelected].langCode}
                                        height="calc(100% - 10px)"
                                        width="calc(100% - 25px)"
                                        value={codeExample.producer}
                                    />
                                    <Copy data={codeExample.producer} />
                                </div>
                            </div>
                        )}

                        {tabValue === 'Consumer' && (
                            <div className="code-example">
                                <div className="code-content">
                                    <Editor
                                        options={{
                                            minimap: { enabled: false },
                                            scrollbar: { verticalScrollbarSize: 0 },
                                            scrollBeyondLastLine: false,
                                            roundedSelection: false,
                                            formatOnPaste: true,
                                            formatOnType: true,
                                            readOnly: true,
                                            fontSize: '14px',
                                            fontFamily: 'Inter'
                                        }}
                                        language={SDK_CODE_EXAMPLE[langSelected].langCode}
                                        height="calc(100% - 10px)"
                                        width="calc(100% - 25px)"
                                        value={codeExample.consumer}
                                    />
                                    <Copy data={codeExample.consumer} />
                                </div>
                            </div>
                        )}
                    </div>
                </>
            )}
            {protocolSelected === 'SDK(TCP)' && (
                <>
                    <div className="installation">
                        <p className="field-title">First, generate a token</p>
                        <div className="code-example">
                            <div className="code-content">
                                <Editor
                                    options={{
                                        minimap: { enabled: false },
                                        scrollbar: { verticalScrollbarSize: 0 },
                                        scrollBeyondLastLine: false,
                                        roundedSelection: false,
                                        formatOnPaste: true,
                                        formatOnType: true,
                                        readOnly: true,
                                        fontSize: '14px',
                                        fontFamily: 'Inter'
                                    }}
                                    language={PROTOCOL_CODE_EXAMPLE[langSelected].langCode}
                                    height="calc(100% - 10px)"
                                    width="calc(100% - 25px)"
                                    value={codeExample.tokenGenerate}
                                />
                                <Copy data={codeExample.tokenGenerate} />
                            </div>
                        </div>
                    </div>
                    <div className="tabs">
                        {/* {showTabs && <CustomTabs value={tabValue} onChange={(tabValue) => setTabValue(tabValue)} tabs={tabs}></CustomTabs>} */}
                        {/* {tabValue === 'Producer' && ( */}
                        <p className="field-title">Producer</p>

                        <div className="code-example">
                            <div className="code-content produce">
                                <Editor
                                    options={{
                                        minimap: { enabled: false },
                                        scrollbar: { verticalScrollbarSize: 0 },
                                        scrollBeyondLastLine: false,
                                        roundedSelection: false,
                                        formatOnPaste: true,
                                        formatOnType: true,
                                        readOnly: true,
                                        fontSize: '14px',
                                        fontFamily: 'Inter'
                                    }}
                                    language={PROTOCOL_CODE_EXAMPLE[langSelected].langCode}
                                    height="calc(100% - 10px)"
                                    width="calc(100% - 25px)"
                                    value={codeExample.producer}
                                />
                                <Copy data={codeExample.producer} />
                            </div>
                        </div>
                        {/* )} */}
                    </div>
                </>
            )}
        </div>
    );
};

export default SdkExample;
