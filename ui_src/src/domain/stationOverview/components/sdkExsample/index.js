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

import './style.scss';

import React, { useEffect, useState } from 'react';

import SelectComponent from '../../../../components/select';
import { CODE_EXAMPLE } from '../../../../const/SDKExample';
import { LOCAL_STORAGE_ENV, LOCAL_STORAGE_NAMESPACE } from '../../../../const/localStorageConsts';
import CustomTabs from '../../../../components/Tabs';
import Copy from '../../../../components/copy';
import * as monaco from 'monaco-editor';
import Editor from '@monaco-editor/react';
import { loader } from '@monaco-editor/react';
loader.config({ monaco });

const tabs = ['Producer', 'Consumer'];

const SdkExample = ({ consumer, showTabs = true }) => {
    const [langSelected, setLangSelected] = useState('Go');
    const selectLngOption = ['Go', 'Node.js', 'Typescript', 'Python'];
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
        codeEx.producer = CODE_EXAMPLE[lang].producer;
        codeEx.consumer = CODE_EXAMPLE[lang].consumer;
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

    useEffect(() => {
        changeDynamicCode(langSelected);
    }, []);

    const handleSelectLang = (e) => {
        setLangSelected(e);
        changeDynamicCode(e);
    };

    return (
        <div className="sdk-details-container">
            <div className="select-lan">
                <p>Language</p>
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
            <div className="installation">
                <p>Package installation</p>
                <div className="install-copy">
                    <p>{CODE_EXAMPLE[langSelected].installation}</p>
                    <Copy data={CODE_EXAMPLE[langSelected].installation} />
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
                                    fontSize: '14px'
                                }}
                                language={CODE_EXAMPLE[langSelected].langCode}
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
                                    fontSize: '14px'
                                }}
                                language={CODE_EXAMPLE[langSelected].langCode}
                                height="calc(100% - 10px)"
                                width="calc(100% - 25px)"
                                value={codeExample.consumer}
                            />
                            <Copy data={codeExample.consumer} />
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
};

export default SdkExample;
