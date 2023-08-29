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
import { ReactComponent as TotalMsgIcon } from '../../../../assets/images/setting/totalMsgIcon.svg';
import { ReactComponent as PriceIcon } from '../../../../assets/images/setting/priceIcon.svg';
import { ReactComponent as RequestsInIcon } from '../../../../assets/images/setting/requestsIn.svg';
import { ReactComponent as RequestsOutIcon } from '../../../../assets/images/setting/requestsOut.svg';
import { ReactComponent as CloudProviderAWSIcon } from '../../../../assets/images/setting/cloudProviderAWS.svg';
import { ReactComponent as USAIcon } from '../../../../assets/images/setting/usaIcon.svg';
import { ReactComponent as GermanyIcon } from '../../../../assets/images/setting/germanyIcon.svg';
import { ReactComponent as ConsumedIcon } from '../../../../assets/images/setting/consumed.svg';
import { ReactComponent as RedeliverIcon } from '../../../../assets/images/setting/redeliver.svg';
import { ReactComponent as DeadLetterIcon } from '../../../../assets/images/setting/deadLetter.svg';
import { ReactComponent as StorageIcon } from '../../../../assets/images/setting/storage.svg';
import DatePickerComponent from '../../../../components/datePicker';
import SegmentButton from '../../../../components/segmentButton';
import Loader from '../../../../components/loader';
import { LOCAL_STORAGE_CREATION_DATE } from '../../../../const/localStorageConsts';
function Requests() {
    const [usageData, setUsageData] = useState(null);
    const [usageType, setUsageType] = useState('Data out');
    const [isLoading, setIsLoading] = useState(true);
    const [totalDiscount, setTotalDiscount] = useState(0);
    const [totalDataInPrice, setTotalDataInPrice] = useState(0);
    const [totalDataOutPrice, setTotalDataOutPrice] = useState(0);
    const [displayMonth, setDisplayMonth] = useState();

    const getBillingDetails = async (date) => {
        try {
            const month = date.getMonth();
            const year = date.getFullYear();
            const data = await httpRequest('GET', `${ApiEndpoints.GET_BILLING_DETAILS}?month=${month + 1}&year=${year}`);
            setTotalDiscount(data?.total_free_tier_discount + data?.discount);
            setTotalDataInPrice(data?.price_per_gb_in * convertBytesToGb(data?.data_in));
            setTotalDataOutPrice(data?.price_per_gb_out * convertBytesToGb(data?.data_out));
            setUsageData(data);
            setIsLoading(false);
        } catch (err) {
            return;
        }
    };

    const formatNumber = (number) => {
        const decimalPlaces = (number.toString().split('.')[1] || '').length;
        switch (decimalPlaces) {
            case decimalPlaces >= 3:
                return number.toFixed(3);
            case decimalPlaces >= 2:
                return number.toFixed(2);
            case decimalPlaces >= 1:
                return number.toFixed(1);
            default:
                return number;
        }
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
            return `Next billing date is ${getNextPaymentDate()}. \nPrice will be calculated based on the usage of this month.`;
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
        } else return GermanyIcon;
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
                    <DatePickerComponent onChange={onChangeDate} picker="month" allowClear={false} dateFrom={localStorage.getItem(LOCAL_STORAGE_CREATION_DATE)} />
                </div>
            </div>
            <div className="usage-header-section">
                <div className="requests-summary">
                    <div className="requests-summary-in-out">
                        <div className="data-in">
                            <div className="requests-total">
                                <RequestsInIcon alt="data in" />
                                <span className="requests-data">
                                    <label className="requests-title-in">Data in</label>
                                    <label className="data-gb">{usageData && formatNumber(convertBytesToGb(usageData?.data_in))?.toLocaleString('en-US')}Gb</label>
                                </span>
                            </div>
                            <div className="total-messages">
                                <div className="total-messages-in">
                                    <TotalMsgIcon alt="data in" />
                                    <span className="requests-data">
                                        <label className="requests-title-in">Total Events</label>

                                        <label className="total-value">{usageData ? usageData?.data_in_events?.toLocaleString('en-US') : 0}</label>
                                    </span>
                                </div>
                                <div className="total-messages-in">
                                    <PriceIcon alt="data in" />
                                    <span className="requests-data">
                                        <label className="requests-title-in">Price Per Gb</label>
                                        <label className="total-value">${usageData && usageData?.price_per_gb_in?.toFixed(2).toLocaleString('en-US')}</label>
                                    </span>
                                </div>
                            </div>
                            <span className="cloud-provider">
                                <label className="cloud-provider-label">Provider: </label> <CloudProviderAWSIcon alt="cloud provider" />
                            </span>
                        </div>
                        <div className="data-out">
                            <div className="requests-total">
                                <RequestsOutIcon alt="data out" />
                                <span className="requests-data">
                                    <label className="requests-title-out">Data out</label>
                                    <label className="data-gb">{usageData && formatNumber(convertBytesToGb(usageData?.data_out))?.toLocaleString('en-US')}Gb</label>
                                </span>
                            </div>
                            <div className="total-messages">
                                <div className="total-messages-out">
                                    <TotalMsgIcon alt="data out" />
                                    <span className="requests-data">
                                        <label className="requests-title-in">Total Events</label>
                                        <label className="total-value">{usageData ? usageData?.data_out_events?.toLocaleString('en-US') : 0}</label>
                                    </span>
                                </div>
                                <div className="total-messages-out">
                                    <PriceIcon alt="data out" />
                                    <span className="requests-data">
                                        <label className="requests-title-in">Price Per Gb</label>
                                        <label className="total-value">${usageData && usageData?.price_per_gb_out?.toFixed(2).toLocaleString('en-US')}</label>
                                    </span>
                                </div>
                            </div>
                            <span className="cloud-provider">
                                <label className="cloud-provider-label">Region: </label> <img src={getRegionImage(usageData?.region)} alt="region" />
                                <label className="region">{usageData?.region === '' ? 'eu-central-1' : usageData?.region}</label>
                            </span>
                        </div>
                    </div>
                </div>
                <div className="total-payment">
                    <div className="total-payment-header">
                        <span>
                            <p className="total-ammount">Total Payment</p>
                            <p className="next-billing">{displayMonth && genetrateSentence()}</p>
                        </span>
                        <span className="price-val-star">
                            <label className="requests-value">${usageData?.total_price_after_discount?.toLocaleString('en-US')}</label>
                            <p className="pricing-disclaimer">*</p>
                        </span>
                    </div>
                    <Divider />
                    <span className="billing-item">
                        <p className="item">Data in</p>
                        <p className="ammount">${totalDataInPrice?.toLocaleString('en-US')}</p>
                    </span>
                    <span className="billing-item">
                        <p className="item">Data out</p>
                        <p className="ammount">${totalDataOutPrice?.toLocaleString('en-US')}</p>
                    </span>
                    <span className="billing-item">
                        <p className="item">Discount</p>
                        <p className="ammount">${totalDiscount?.toLocaleString('en-US')}</p>
                    </span>
                    <Divider />
                    <span className="billing-item">
                        <p className="item">Total price</p>
                        <p className="ammount">${usageData?.total_price_after_discount?.toLocaleString('en-US')}</p>
                    </span>
                    <span className="billing-item">
                        <p className="pricing-disclaimer">*Please note that the pricing is not final</p>
                    </span>
                </div>
            </div>
            <div className="usage-details">
                <div className="segment-data">
                    <SegmentButton size="medium" value={usageType} options={['Data in', 'Data out']} onChange={(e) => setUsageType(e)} />
                </div>
                {usageType === 'Data out' && (
                    <div className="panel-container">
                        <div className="requests-panel">
                            <div className="requests-item">
                                <div className="box-edge yellow"></div>
                                <div className="circle-img">
                                    <ConsumedIcon alt="Consumed" />
                                </div>

                                <div>
                                    <label className="request-type">Consumed</label>
                                    <label className="request-description">The total number of consumed events.</label>
                                </div>
                            </div>
                            <label className="requests-value">{usageData?.consumed_events?.toLocaleString('en-US')}</label>
                        </div>
                        <div className="requests-panel">
                            <div className="requests-item">
                                <div className="box-edge yellow"></div>
                                <div className="circle-img">
                                    <RedeliverIcon alt="Redelivered" />
                                </div>

                                <div>
                                    <label className="request-type">Redelivered</label>
                                    <label className="request-description">The total number of redelivered events.</label>
                                </div>
                            </div>
                            <label className="requests-value">{usageData?.redelivery_events?.toLocaleString('en-US')}</label>
                        </div>
                        <div className="requests-panel">
                            <div className="requests-item">
                                <div className="box-edge yellow"></div>
                                <div className="circle-img">
                                    <StorageIcon alt="Storage" />
                                </div>

                                <div>
                                    <label className="request-type">Storage tiering</label>
                                    <label className="request-description">The total number of events migrated using storage tiering.</label>
                                </div>
                            </div>
                            <label className="requests-value">{usageData?.storage_tiering_events?.toLocaleString('en-US')}</label>
                        </div>
                        <div className="requests-panel">
                            <div className="requests-item">
                                <div className="box-edge yellow"></div>
                                <div className="circle-img">
                                    <DeadLetterIcon alt="Dead-letter" />
                                </div>

                                <div>
                                    <label className="request-type">Dead-letter events retrieval</label>
                                    <label className="request-description">The total number of events retransmitted from Dead-Letter Stations.</label>
                                </div>
                            </div>
                            <label className="requests-value">{usageData?.dls_retransmit_events?.toLocaleString('en-US')}</label>
                        </div>
                    </div>
                )}
                {usageType === 'Data in' && (
                    <div className="panel-container">
                        <div className="requests-panel">
                            <div className="requests-item">
                                <div className="box-edge green"></div>
                                <div className="circle-img">
                                    <ConsumedIcon alt="Consumed" />
                                </div>

                                <div>
                                    <label className="request-type">Data in</label>
                                    <label className="request-description">The total number of produced events.</label>
                                </div>
                            </div>
                            <label className="requests-value">{usageData?.data_in_events?.toLocaleString('en-US')}</label>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}
export default Requests;
