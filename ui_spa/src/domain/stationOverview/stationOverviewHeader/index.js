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

import './style.scss';

import React, { useContext, useEffect, useState } from 'react';
import { CopyBlock, atomOneLight } from 'react-code-blocks';
import { useHistory } from 'react-router-dom';
import { Progress } from 'antd';

import { CODE_EXAMPLE, DOCKER_CODE_EXAMPLE } from '../../../const/SDKExample';
import { convertBytes, convertSecondsToDate } from '../../../services/valueConvertor';
import averageMesIcon from '../../../assets/images/averageMesIcon.svg';
import awaitingIcon from '../../../assets/images/awaitingIcon.svg';
import storageIcon from '../../../assets/images/storageIcon.svg';
import memoryIcon from '../../../assets/images/memoryIcon.svg';
import HealthyBadge from '../../../components/healthyBadge';
import cpuIcon from '../../../assets/images/cpuIcon.svg';
import SelectComponent from '../../../components/select';
import Button from '../../../components/button';
import { Context } from '../../../hooks/store';
import Modal from '../../../components/modal';
import pathDomains from '../../../router';
import { StationStoreContext } from '..';
import TooltipComponent from '../../../components/tooltip/tooltip';
import Auditing from '../auditing';
import { InfoOutlined } from '@material-ui/icons';
import { LOCAL_STORAGE_ENV, LOCAL_STORAGE_NAMESPACE } from '../../../const/localStorageConsts';
import CustomTabs from '../../../components/Tabs';

const StationOverviewHeader = (props) => {
    const [state, dispatch] = useContext(Context);
    const [stationState, stationDispatch] = useContext(StationStoreContext);
    const history = useHistory();
    const [retentionValue, setRetentionValue] = useState('');
    const [sdkModal, setSdkModal] = useState(false);
    const [auditModal, setAuditModal] = useState(false);
    const selectLngOption = ['Go', 'Node.js', 'Typescript', 'Python'];
    const [langSelected, setLangSelected] = useState('Go');
    const [codeExample, setCodeExample] = useState({
        import: '',
        connect: '',
        producer: '',
        consumer: ''
    });
    const [tabValue, setTabValue] = useState('0');
    const tabs = ['Producer', 'Consumer'];
    const url = window.location.href;
    const stationName = url.split('factories/')[1].split('/')[1];
    const handleSelectLang = (e) => {
        setLangSelected(e);
        changeDynamicCode(e);
    };
    useEffect(() => {
        switch (stationState?.stationMetaData?.retention_type) {
            case 'message_age_sec':
                setRetentionValue(convertSecondsToDate(stationState?.stationMetaData?.retention_value));
                break;
            case 'bytes':
                setRetentionValue(`${stationState?.stationMetaData?.retention_value} bytes`);
                break;
            case 'messages':
                setRetentionValue(`${stationState?.stationMetaData?.retention_value} messages`);
                break;
            default:
                break;
        }
    }, [stationState?.stationMetaData?.retention_type]);

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
        codeEx.producer = codeEx.producer.replaceAll('<station_name>', stationName);
        codeEx.consumer = codeEx.consumer.replaceAll('<station_name>', stationName);
        setCodeExample(codeEx);
    };

    const returnToStaionsList = () => {
        const url = window.location.href;
        const staionName = url.split('factories/')[1].split('/')[0];
        history.push(`${pathDomains.factoriesList}/${staionName}`);
    };

    return (
        <div className="station-overview-header">
            <div className="title-wrapper">
                <div className="station-details">
                    <h1 className="station-name">{stationState?.stationMetaData?.name}</h1>
                    <span className="created-by">
                        Created by {stationState?.stationMetaData?.created_by_user} at {stationState?.stationMetaData?.creation_date}
                    </span>
                </div>
                <div id="e2e-tests-station-close-btn">
                    <Button
                        width="80px"
                        height="32px"
                        placeholder="Back"
                        colorType="white"
                        radiusType="circle"
                        backgroundColorType="navy"
                        fontSize="13px"
                        fontWeight="600"
                        border="navy"
                        onClick={() => returnToStaionsList()}
                    />
                </div>
            </div>
            <div className="details">
                <div className="main-details">
                    <p>
                        <b>Retention:</b> {retentionValue}
                    </p>
                    <p>
                        <b>Replicas:</b> {stationState?.stationMetaData?.replicas}
                    </p>
                    <p>
                        <b>Storage Type:</b> {stationState?.stationMetaData?.storage_type}
                    </p>
                </div>
                <div className="icons-wrapper">
                    <div className="details-wrapper">
                        <div className="icon">
                            <img src={awaitingIcon} width={22} height={44} alt="awaitingIcon" />
                        </div>
                        <div className="more-details">
                            <p className="title">Total messages</p>
                            <p className="number">{stationState?.stationSocketData?.total_messages || 0}</p>
                        </div>
                    </div>
                    <TooltipComponent text="Include extra bytes added by memphis." width={'220px'} cursor="pointer">
                        <div className="details-wrapper average">
                            <div className="icon">
                                <img src={averageMesIcon} width={24} height={24} alt="averageMesIcon" />
                            </div>
                            <div className="more-details">
                                <p className="title">Av. message size</p>
                                <p className="number">{convertBytes(stationState?.stationSocketData?.average_message_size)}</p>
                            </div>
                        </div>
                    </TooltipComponent>
                    {/* <div className="details-wrapper">
                        <div className="icon">
                            <img src={memoryIcon} width={24} height={24} alt="memoryIcon" />
                        </div>
                        <div className="more-details">
                            <p className="number">20Mb/80Mb</p>
                            <Progress showInfo={false} status={(20 / 80) * 100 > 60 ? 'exception' : 'success'} percent={(20 / 80) * 100} size="small" />
                            <p className="title">Mem</p>
                        </div>
                    </div>
                    <div className="details-wrapper">
                        <div className="icon">
                            <img src={cpuIcon} width={22} height={22} alt="cpuIcon" />
                        </div>
                        <div className="more-details">
                            <p className="number">50%</p>
                            <Progress showInfo={false} status={(35 / 100) * 100 > 60 ? 'exception' : 'success'} percent={(35 / 100) * 100} size="small" />
                            <p className="title">CPU</p>
                        </div>
                    </div>
                    <div className="details-wrapper">
                        <div className="icon">
                            <img src={storageIcon} width={30} height={30} alt="storageIcon" />
                        </div>
                        <div className="more-details">
                            <p className="number">{60}Mb/100Mb</p>
                            <Progress showInfo={false} status={(60 / 100) * 100 > 60 ? 'exception' : 'success'} percent={(60 / 100) * 100} size="small" />
                            <p className="title">Storage</p>
                        </div>
                    </div> */}
                </div>
                <div className="info-buttons">
                    <div className="sdk">
                        <p>SDK</p>
                        <span
                            onClick={() => {
                                changeDynamicCode(langSelected);
                                setSdkModal(true);
                            }}
                        >
                            View Details {'>'}
                        </span>
                    </div>
                    <div className="audit">
                        <p>Audit</p>
                        <span onClick={() => setAuditModal(true)}>View Details {'>'}</span>
                    </div>
                </div>
            </div>
            <Modal header="SDK" width="710px" clickOutside={() => setSdkModal(false)} open={sdkModal} displayButtons={false}>
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
                            dropdownClassName="select-options"
                        />
                    </div>
                    <div className="installation">
                        <p>Installation</p>
                        <div className="install-copy">
                            <p></p>
                            <CopyBlock text={CODE_EXAMPLE[langSelected].installation} showLineNumbers={false} theme={atomOneLight} wrapLines={true} codeBlock />
                        </div>
                    </div>
                    <div className="tabs">
                        <CustomTabs value={tabValue} onChange={(tabValue) => setTabValue(tabValue)} tabs={tabs}></CustomTabs>
                        {tabValue === '0' && (
                            <div className="code-example">
                                <div className="code-content">
                                    <CopyBlock
                                        language={CODE_EXAMPLE[langSelected].langCode}
                                        text={codeExample.producer}
                                        showLineNumbers={true}
                                        theme={atomOneLight}
                                        wrapLines={true}
                                        codeBlock
                                    />
                                </div>
                            </div>
                        )}

                        {tabValue === '1' && (
                            <div className="code-example">
                                <div className="code-content">
                                    <CopyBlock
                                        language={CODE_EXAMPLE[langSelected].langCode}
                                        text={codeExample.consumer}
                                        showLineNumbers={true}
                                        theme={atomOneLight}
                                        wrapLines={true}
                                        codeBlock
                                    />
                                </div>
                            </div>
                        )}
                    </div>
                </div>
            </Modal>
            <Modal
                header={
                    <div className="audit-header">
                        <p className="title">Audit</p>
                        <div className="msg">
                            <InfoOutlined />
                            <p>Showing last 5 days</p>
                        </div>
                    </div>
                }
                displayButtons={false}
                height="300px"
                width="800px"
                clickOutside={() => setAuditModal(false)}
                open={auditModal}
                hr={false}
            >
                <Auditing />
            </Modal>
        </div>
    );
};

export default StationOverviewHeader;
