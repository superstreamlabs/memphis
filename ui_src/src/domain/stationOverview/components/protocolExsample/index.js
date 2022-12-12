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
import { PROTOCOL_CODE_EXAMPLE } from '../../../../const/codeExample';
import { LOCAL_STORAGE_ENV, LOCAL_STORAGE_NAMESPACE } from '../../../../const/localStorageConsts';
import CustomTabs from '../../../../components/Tabs';
import Copy from '../../../../components/copy';
import Editor from '@monaco-editor/react';

const tabs = ['Producer'];

const ProtocolExample = ({ consumer, showTabs = true }) => {
    const [langSelected, setLangSelected] = useState('Rest');
    const selectLngOption = ['Rest'];
    const [codeExample, setCodeExample] = useState({
        producer: '',
        tokenGenerate: '',
        langCode: ''
    });
    const [tabValue, setTabValue] = useState(consumer ? 'Consumer' : 'Producer');

    const changeDynamicCode = (lang) => {
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

    return (
        <div className="code-example-details-container protocol-example">
            <div className="select-lan">
                <p>Protocol</p>
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
                <p>Token Generate</p>
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
                                language={PROTOCOL_CODE_EXAMPLE[langSelected].langCode}
                                height="calc(100% - 10px)"
                                width="calc(100% - 25px)"
                                value={codeExample.producer}
                            />
                            <Copy data={codeExample.producer} />
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
};

export default ProtocolExample;
