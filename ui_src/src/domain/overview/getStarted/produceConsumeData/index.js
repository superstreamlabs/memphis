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

import React, { useState, useEffect, useContext } from 'react';
import { loader } from '@monaco-editor/react';
import Editor from '@monaco-editor/react';
import * as monaco from 'monaco-editor';

import { LOCAL_STORAGE_ENV, LOCAL_STORAGE_NAMESPACE } from '../../../../const/localStorageConsts';
import successCons from '../../../../assets/images/stationsIconActive.svg';
import successProd from '../../../../assets/images/dataProduced.svg';
import TitleComponent from '../../../../components/titleComponent';
import { SDK_CODE_EXAMPLE } from '../../../../const/codeExample';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import SelectComponent from '../../../../components/select';
import { httpRequest } from '../../../../services/http';
import Button from '../../../../components/button';
import Copy from '../../../../components/copy';
import { GetStartedStoreContext } from '..';

loader.config({ monaco });

export const produceConsumeScreenEnum = {
    DATA_SNIPPET: 0,
    DATA_WAITING: 1,
    DATA_RECIEVED: 2
};

const ProduceConsumeData = (props) => {
    const { waitingImage, waitingTitle, successfullTitle, activeData, dataName, displayScreen, screen } = props;
    const [currentPhase, setCurrentPhase] = useState(displayScreen);
    const [getStartedState, getStartedDispatch] = useContext(GetStartedStoreContext);
    const [intervalStationDetails, setintervalStationDetails] = useState();

    const [langSelected, setLangSelected] = useState('Go');
    const [codeExample, setCodeExample] = useState({
        import: '',
        producer: '',
        consumer: ''
    });

    const handleSelectLang = (e) => {
        setLangSelected(e);
        changeDynamicCode(e);
    };

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
        codeEx.producer = codeEx.producer.replaceAll('<application type username>', getStartedState?.username);
        codeEx.consumer = codeEx.consumer.replaceAll('<application type username>', getStartedState?.username);
        codeEx.producer = codeEx.producer.replaceAll('<broker-token>', getStartedState?.connectionCreds);
        codeEx.consumer = codeEx.consumer.replaceAll('<broker-token>', getStartedState?.connectionCreds);
        codeEx.producer = codeEx.producer.replaceAll('<station-name>', getStartedState?.stationName);
        codeEx.consumer = codeEx.consumer.replaceAll('<station-name>', getStartedState?.stationName);
        setCodeExample(codeEx);
    };

    const getStationDetails = async () => {
        try {
            const data = await httpRequest('GET', `${ApiEndpoints.GET_STATION_DATA}?station_name=${getStartedState?.formFieldsCreateStation?.name}`);
            if (data) {
                if (data?.messages?.length > 0) {
                    setCurrentPhase(produceConsumeScreenEnum['DATA_RECIEVED']);
                    getStartedDispatch({ type: 'GET_STATION_DATA', payload: data });
                    clearInterval(intervalStationDetails);
                }
            }
        } catch (error) {
            if (error?.status === 666) {
                clearInterval(intervalStationDetails);
            }
        }
    };

    useEffect(() => {
        changeDynamicCode(langSelected);
        if (displayScreen !== currentPhase) {
            if (displayScreen === produceConsumeScreenEnum['DATA_SNIPPET'] || displayScreen === produceConsumeScreenEnum['DATA_WAITING']) {
                onCopyToClipBoard();
            }
            setCurrentPhase(displayScreen);
        }
    }, [displayScreen]);

    const onCopyToClipBoard = () => {
        let interval = setInterval(() => {
            getStationDetails();
        }, 3000);
        setintervalStationDetails(interval);
    };

    useEffect(() => {
        if (currentPhase === produceConsumeScreenEnum['DATA_WAITING']) {
            getStartedDispatch({ type: 'SET_HIDDEN_BUTTON', payload: true });
        } else {
            getStartedDispatch({ type: 'SET_HIDDEN_BUTTON', payload: false });
            getStartedDispatch({ type: 'SET_NEXT_DISABLE', payload: false });
        }

        if (currentPhase === produceConsumeScreenEnum['DATA_RECIEVED']) {
            clearInterval(intervalStationDetails);
        }
        return () => {
            getStartedDispatch({ type: 'SET_HIDDEN_BUTTON', payload: false });
            clearInterval(intervalStationDetails);
        };
    }, [currentPhase]);

    useEffect(() => {
        if (
            getStartedState?.stationData &&
            getStartedState?.stationData[activeData] &&
            Object.keys(getStartedState?.stationData[activeData]).length >= 1 &&
            getStartedState?.stationData[activeData][0]?.name === dataName
        ) {
            setCurrentPhase(produceConsumeScreenEnum['DATA_RECIEVED']);
        }
        return () => {
            if (currentPhase === produceConsumeScreenEnum['DATA_RECIEVED']) {
                clearInterval(intervalStationDetails);
            }
        };
    }, [[getStartedState?.stationData?.[activeData]]]);

    return (
        <div className="create-produce-data">
            {currentPhase === produceConsumeScreenEnum['DATA_SNIPPET'] && (
                <div className="code-snippet">
                    <div className="lang">
                        <p className="title">Language</p>
                        <SelectComponent
                            initialValue={langSelected}
                            value={langSelected}
                            colorType="navy"
                            backgroundColorType="none"
                            borderColorType="gray"
                            radiusType="semi-round"
                            width="250px"
                            height="50px"
                            options={props.languages}
                            onChange={(e) => handleSelectLang(e)}
                            popupClassName="select-options"
                        />
                    </div>
                    <div className="installation">
                        <p className="title">Installation</p>
                        <div className="install-copy">
                            <p>{SDK_CODE_EXAMPLE[langSelected].installation}</p>
                            <Copy data={SDK_CODE_EXAMPLE[langSelected].installation} />
                        </div>
                    </div>
                    <div className="code-example">
                        {props.produce ? (
                            <div>
                                <p className="title">Code snippet for producing data</p>
                                {/* <p className="description">Just copy and paste the following code to your preferred IDE</p> */}
                            </div>
                        ) : (
                            <div>
                                <p className="title">Code snippet for consuming data</p>
                                {/* <p className="description">Just copy and paste the following code to your preferred IDE</p> */}
                            </div>
                        )}
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
                                value={props.produce ? codeExample.producer : codeExample.consumer}
                            />
                            <Copy data={props.produce ? codeExample.producer : codeExample.consumer} />
                        </div>
                    </div>
                </div>
            )}
            {currentPhase === produceConsumeScreenEnum['DATA_WAITING'] && (
                <div className="code-snippet">
                    <div className="lang">
                        <p className="title">Language</p>
                        <SelectComponent
                            initialValue={langSelected}
                            value={langSelected}
                            colorType="navy"
                            backgroundColorType="none"
                            borderColorType="gray"
                            radiusType="semi-round"
                            width="250px"
                            height="50px"
                            options={props.languages}
                            onChange={(e) => handleSelectLang(e)}
                            dropdownClassName="select-options"
                            disabled
                        />
                    </div>
                    <div className="installation">
                        <p className="title">Installation</p>
                        <div className="install-copy">
                            <p>{SDK_CODE_EXAMPLE[langSelected].installation}</p>
                            <Copy data={SDK_CODE_EXAMPLE[langSelected].installation} />
                        </div>
                    </div>
                    <div className="data-waiting-container">
                        <img className="image-waiting-successful" src={waitingImage} alt={'waitingImage'} />
                        <TitleComponent headerTitle={waitingTitle} typeTitle="sub-header" style={{ header: { fontSize: '18px' } }}></TitleComponent>
                        <div className="waiting-for-data-btn">
                            <Button
                                width="129px"
                                height="40px"
                                placeholder="Back"
                                colorType="white"
                                radiusType="circle"
                                backgroundColorType="black"
                                border="gray"
                                fontSize="14px"
                                fontWeight="bold"
                                marginBottom="3px"
                                onClick={() => {
                                    clearInterval(intervalStationDetails);
                                    screen(produceConsumeScreenEnum['DATA_SNIPPET']);
                                }}
                            />
                            <div className="waiting-for-data-space"></div>
                            <div>
                                <Button
                                    width="129px"
                                    height="40px"
                                    placeholder="Skip"
                                    colorType="black"
                                    radiusType="circle"
                                    backgroundColorType="white"
                                    border="gray"
                                    fontSize="14px"
                                    fontWeight="bold"
                                    marginBottom="3px"
                                    onClick={() => {
                                        clearInterval(intervalStationDetails);
                                        getStartedDispatch({ type: 'SET_COMPLETED_STEPS', payload: getStartedState?.currentStep });
                                        getStartedDispatch({ type: 'SET_CURRENT_STEP', payload: getStartedState?.currentStep + 1 });
                                    }}
                                />
                            </div>
                        </div>
                    </div>
                </div>
            )}
            {currentPhase === produceConsumeScreenEnum['DATA_RECIEVED'] && (
                <div className="code-snippet">
                    <div className="lang">
                        <p className="title">Language</p>
                        <SelectComponent
                            initialValue={langSelected}
                            value={langSelected}
                            colorType="navy"
                            backgroundColorType="none"
                            borderColorType="gray"
                            radiusType="semi-round"
                            width="250px"
                            height="50px"
                            options={props.languages}
                            onChange={(e) => handleSelectLang(e)}
                            dropdownClassName="select-options"
                            disabled
                        />
                    </div>
                    <div className="installation">
                        <p className="title">Installation</p>
                        <div className="install-copy">
                            <p>{SDK_CODE_EXAMPLE[langSelected].installation}</p>
                            <Copy data={SDK_CODE_EXAMPLE[langSelected].installation} />
                        </div>
                    </div>
                    <div className="successfully-container">
                        {props.produce ? (
                            <img className="image-waiting-successful" src={successProd} alt="successProd" />
                        ) : (
                            <img className="image-waiting-successful" src={successCons} alt="successCons" />
                        )}

                        <TitleComponent headerTitle={successfullTitle} typeTitle="sub-header" style={{ header: { fontSize: '18px' } }}></TitleComponent>
                    </div>
                </div>
            )}
        </div>
    );
};

export default ProduceConsumeData;
