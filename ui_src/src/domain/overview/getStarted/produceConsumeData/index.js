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

import React, { useState, useEffect, useContext } from 'react';
import Lottie from 'lottie-react';
import { CopyBlock, atomOneLight } from 'react-code-blocks';
import SelectComponent from '../../../../components/select';
import Button from '../../../../components/button';
import successProdCons from '../../../../assets/lotties/successProdCons.json';
import { GetStartedStoreContext } from '..';
import { httpRequest } from '../../../../services/http';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import './style.scss';
import TitleComponent from '../../../../components/titleComponent';
import { CODE_EXAMPLE } from '../../../../const/SDKExample';
import { LOCAL_STORAGE_ENV, LOCAL_STORAGE_NAMESPACE } from '../../../../const/localStorageConsts';

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
        codeEx.producer = CODE_EXAMPLE[lang].producer;
        codeEx.consumer = CODE_EXAMPLE[lang].consumer;
        let host = process.env.REACT_APP_SANDBOX_ENV
            ? 'broker.sandbox.memphis.dev'
            : localStorage.getItem(LOCAL_STORAGE_ENV) === 'docker'
            ? 'localhost'
            : 'memphis-cluster.' + localStorage.getItem(LOCAL_STORAGE_NAMESPACE) + '.svc.cluster.local';
        codeEx.producer = codeEx.producer.replaceAll('<memphis-host>', host);
        codeEx.consumer = codeEx.consumer.replaceAll('<memphis-host>', host);
        codeEx.producer = codeEx.producer.replaceAll('<application type username>', getStartedState?.username);
        codeEx.consumer = codeEx.consumer.replaceAll('<application type username>', getStartedState?.username);
        codeEx.producer = codeEx.producer.replaceAll('<connection_token>', getStartedState?.connectionCreds);
        codeEx.consumer = codeEx.consumer.replaceAll('<connection_token>', getStartedState?.connectionCreds);
        codeEx.producer = codeEx.producer.replaceAll('<station_name>', getStartedState?.stationName);
        codeEx.consumer = codeEx.consumer.replaceAll('<station_name>', getStartedState?.stationName);
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
            if (displayScreen === produceConsumeScreenEnum['DATA_WAITING']) {
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
                        <p>Language</p>
                        <SelectComponent
                            initialValue={langSelected}
                            value={langSelected}
                            colorType="navy"
                            backgroundColorType="none"
                            borderColorType="gray"
                            radiusType="semi-round"
                            width="450px"
                            height="50px"
                            options={props.languages}
                            onChange={(e) => handleSelectLang(e)}
                            dropdownClassName="select-options"
                        />
                    </div>
                    <div className="installation">
                        <p>Installation</p>
                        <div className="install-copy">
                            <CopyBlock text={CODE_EXAMPLE[langSelected].installation} showLineNumbers={false} theme={atomOneLight} wrapLines={true} codeBlock />
                        </div>
                    </div>
                    <div className="code-example">
                        <p>{props.produce ? 'Produce data' : 'Consume data'}</p>
                        <div className="code-content">
                            <CopyBlock
                                language={CODE_EXAMPLE[langSelected].langCode}
                                text={props.produce ? codeExample.producer : codeExample.consumer}
                                showLineNumbers={true}
                                theme={atomOneLight}
                                wrapLines={true}
                                codeBlock
                            />
                        </div>
                    </div>
                </div>
            )}
            {currentPhase === produceConsumeScreenEnum['DATA_WAITING'] && (
                <div className="data-waiting-container">
                    <Lottie className="image-waiting-successful" animationData={waitingImage} loop={true} />
                    <TitleComponent headerTitle={waitingTitle} typeTitle="sub-header" style={{ header: { fontSize: '18px' } }}></TitleComponent>
                    <div className="waiting-for-data-btn">
                        <Button
                            width="129px"
                            height="40px"
                            placeholder="Back"
                            colorType="black"
                            radiusType="circle"
                            backgroundColorType="white"
                            border="border: 1px solid #EBEBEB"
                            fontSize="14px"
                            fontWeight="bold"
                            marginBottom="3px"
                            onClick={() => {
                                clearInterval(intervalStationDetails);
                                screen(produceConsumeScreenEnum['DATA_SNIPPET']);
                            }}
                        />
                        <div className="waiting-for-data-space"></div>
                        <div id="e2e-getstarted-skip">
                            <Button
                                width="129px"
                                height="40px"
                                placeholder="Skip"
                                colorType="black"
                                radiusType="circle"
                                backgroundColorType="white"
                                border="border: 1px solid #EBEBEB"
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
            )}
            {currentPhase === produceConsumeScreenEnum['DATA_RECIEVED'] && (
                <div className="successfully-container">
                    <Lottie className="image-waiting-successful" animationData={successProdCons} loop={true} />
                    <TitleComponent headerTitle={successfullTitle} typeTitle="sub-header" style={{ header: { fontSize: '18px' } }}></TitleComponent>
                </div>
            )}
        </div>
    );
};

export default ProduceConsumeData;
