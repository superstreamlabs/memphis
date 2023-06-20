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
import React, { useEffect, useState } from 'react';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import { httpRequest } from '../../../../services/http';
import { convertBytesToGb } from '../../../../services/valueConvertor';
import { Divider } from 'antd';
import TotalRequests from '../../../../assets/images/setting/totalRequests.svg';
import RequestsIn from '../../../../assets/images/setting/requestsIn.svg';
import RequestsOut from '../../../../assets/images/setting/requestsOut.svg';
import CloudProviderAWS from '../../../../assets/images/setting/cloudProviderAWS.svg';
import USAIcon from '../../../../assets/images/setting/usaIcon.svg';
import GermanyIcon from '../../../../assets/images/setting/germanyIcon.svg';
import AvgMsgSize from '../../../../assets/images/setting/avgMsgSize.svg';
import Consumed from '../../../../assets/images/setting/consumed.svg';
import Redeliver from '../../../../assets/images/setting/redeliver.svg';
import DeadLetter from '../../../../assets/images/setting/deadLetter.svg';
import Storage from '../../../../assets/images/setting/storage.svg';
import DatePickerComponent from '../../../../components/datePicker';
import SegmentButton from '../../../../components/segmentButton';
import Loader from '../../../../components/loader';

function Requests() {
    const [usageData, setUsageData] = useState(null);
    const [usageType, setUsageType] = useState('Data in');
    const [isLoading, setIsLoading] = useState(true);
    const [displayMonth, setDisplayMonth] = useState();

    const getBillingDetails = async (date) => {
        try {
            const month = date.getMonth();
            const year = date.getFullYear();
            const data = await httpRequest('GET', `${ApiEndpoints.GET_BILLING_DETAILS}?month=${month + 1}&year=${year}`);
            setUsageData(data);
            setIsLoading(false);
        } catch (err) {
            return;
        }
    };

    const getTotalEvents = () => {
        return usageData?.data_in_events + usageData?.data_out_events;
    };

    const getNextPaymentDate = () => {
        const today = new Date();
        const nextMonth = ((today.getMonth() + 1) % 12) + 1;
        const year = today.getFullYear() + (nextMonth === 1 ? 1 : 0);
        today.setMonth(nextMonth - 1);
        return `01 ${today.toLocaleString('en-US', { month: 'long' })} ${year}`;
    };

    const genetrateSentence = () => {
        const today = new Date();
        const month = (today.getMonth() + 1) % 12;
        const year = today.getFullYear();
        if (displayMonth.month === month && displayMonth.year === year)
            return `Next billing date is ${getNextPaymentDate()}. Price will be calculated based on the usage of this month.`;
        else {
            const prevDate = new Date(displayMonth.year, displayMonth.month - 1, 1);
            return `Billing details for ${prevDate.toLocaleString('en-US', { month: 'long' })} ${displayMonth.year}`;
        }
    };

    const onChangeDate = (date) => {
        setDisplayMonth({ month: date.getMonth() + 1, year: date.getFullYear() });
        getBillingDetails(date);
    };

    useEffect(() => {
        const today = new Date();
        const month = (today.getMonth() + 1) % 12;
        const year = today.getFullYear();
        setDisplayMonth({ month: month, year: year });
        getBillingDetails(today);
    }, []);

    const getRegionImage = (region) => {
        if (region === 'us-east-1') {
            return USAIcon;
        } else if (region === 'eu-central-1') {
            return GermanyIcon;
        }
    };
    return (
        <div className="requests-container">
            {isLoading && <Loader />}
            <div className="header-preferences">
                <div className="header">
                    <div>
                        <p className="main-header">Usage report</p>
                        <p className="memphis-label">We will keep an eye on your data streams and alert.</p>
                    </div>
                    <DatePickerComponent onChange={onChangeDate} picker="month" allowClear={false} />
                </div>
            </div>
            <div className="usage-header-section">
                <div className="requests-summary">
                    <div className="cloud-provider">
                        <span>
                            <label className="cloud-provider-label">Provider: </label> <img src={CloudProviderAWS} alt="cloud provider" />
                        </span>
                        <Divider type="vertical" />
                        {usageData?.region !== '' && (
                            <span>
                                <label className="cloud-provider-label">Region: </label> <img src={getRegionImage()} alt="region" />
                                <label className="region">{usageData?.region}</label>
                            </span>
                        )}
                    </div>
                    <Divider />
                    <div className="requests-total">
                        <img src={TotalRequests} alt="TotalRequests" />
                        <span className="requests-data">
                            <label className="requests-title">Total requests</label>
                            {usageData && <label className="requests-value">{getTotalEvents().toLocaleString('en-US')}</label>}
                        </span>
                    </div>
                    <Divider />
                    <div className="total-in-out">
                        <div className="requests-total">
                            <img src={RequestsIn} alt="data in" />
                            <span className="requests-data">
                                <label className="requests-title-in">Data in</label>
                                {usageData && <label className="requests-value">{convertBytesToGb(usageData?.data_in_events)?.toLocaleString('en-US')}Gb</label>}{' '}
                            </span>
                        </div>
                        <Divider type="vertical" />
                        <div className="requests-total">
                            <img src={RequestsOut} alt="data out" />
                            <span className="requests-data">
                                <label className="requests-title-out">Data out</label>
                                {usageData && <label className="requests-value">{convertBytesToGb(usageData?.data_out_events)?.toLocaleString('en-US')}Gb</label>}{' '}
                            </span>
                        </div>
                    </div>
                </div>
                <div className="total-payment">
                    <div className="total-payment-header">
                        <span>
                            <p className="total-ammount">Total Payment</p>
                            {/* <p className="next-billing">Next billing date is {getNextPaymentDate()}</p> */}
                            <p className="next-billing">{displayMonth && genetrateSentence()}</p>
                        </span>
                        <label className="requests-value">${usageData?.total_price_after_discount?.toLocaleString('en-US')}</label>
                    </div>
                    <Divider />
                    <span className="billing-item">
                        <p className="item">Total usage</p>
                        <p className="ammount">${usageData?.total_price_before_discount?.toLocaleString('en-US')}</p>
                    </span>
                    <span className="billing-item">
                        <p className="item">Free tier discount</p>
                        <p className="ammount">${usageData?.total_free_tier_discount?.toLocaleString('en-US')}</p>
                    </span>
                    <span className="billing-item">
                        <p className="item">
                            Discount <label className="discount-badge">private-beta</label>
                        </p>
                        <p className="ammount">${usageData?.discount?.toLocaleString('en-US')}</p>
                    </span>
                    <Divider />
                    <span className="billing-item">
                        <p className="item"></p>
                        <p className="ammount">${usageData?.total_price_after_discount?.toLocaleString('en-US')}</p>
                    </span>
                </div>
            </div>
            <div className="usage-details">
                <div className="segment-data">
                    <SegmentButton size="medium" options={['Data in', 'Data out']} onChange={(e) => setUsageType(e)} />
                </div>
                {usageType === 'Data out' && (
                    <div>
                        <div className="requests-panel">
                            <div className="requests-item">
                                <div className="yellow-edge"></div>
                                <div className="circle-img">
                                    <img src={Consumed} alt="Consumed" />
                                </div>

                                <div>
                                    <label className="request-type">Consumed events</label>
                                    <label className="request-description">Contrary to popular belief, Lorem Ipsum</label>
                                </div>
                            </div>
                            <label className="requests-value">{usageData?.consumed_events?.toLocaleString('en-US')}</label>
                        </div>
                        <div className="requests-panel">
                            <div className="requests-item">
                                <div className="yellow-edge"></div>
                                <div className="circle-img">
                                    <img src={Redeliver} alt="Consumed" />
                                </div>

                                <div>
                                    <label className="request-type">Redelivery events</label>
                                    <label className="request-description">Contrary to popular belief, Lorem Ipsum</label>
                                </div>
                            </div>
                            <label className="requests-value">{usageData?.redelivery_events?.toLocaleString('en-US')}</label>
                        </div>
                        <div className="requests-panel">
                            <div className="requests-item">
                                <div className="yellow-edge"></div>
                                <div className="circle-img">
                                    <img src={Storage} alt="Storage" />
                                </div>

                                <div>
                                    <label className="request-type">Storage tiering events</label>
                                    <label className="request-description">Contrary to popular belief, Lorem Ipsum</label>
                                </div>
                            </div>
                            <label className="requests-value">{usageData?.storage_tiering_events?.toLocaleString('en-US')}</label>
                        </div>
                        <div className="requests-panel">
                            <div className="requests-item">
                                <div className="yellow-edge"></div>
                                <div className="circle-img">
                                    <img src={DeadLetter} alt="Consumed" />
                                </div>

                                <div>
                                    <label className="request-type">Dead Letter retransmit events</label>
                                    <label className="request-description">Contrary to popular belief, Lorem Ipsum</label>
                                </div>
                            </div>
                            <label className="requests-value">{usageData?.dls_retransmit_events?.toLocaleString('en-US')}</label>
                        </div>
                    </div>
                )}
                {usageType === 'Data in' && (
                    <div className="requests-panel">
                        <div className="requests-item">
                            <div className="yellow-edge"></div>
                            <div className="circle-img">
                                <img src={Consumed} alt="Consumed" />
                            </div>

                            <div>
                                <label className="request-type">Data in events</label>
                                <label className="request-description">Contrary to popular belief, Lorem Ipsum</label>
                            </div>
                        </div>
                        <label className="requests-value">{usageData?.data_in_events?.toLocaleString('en-US')}</label>
                    </div>
                )}
            </div>
        </div>
    );
}
export default Requests;
