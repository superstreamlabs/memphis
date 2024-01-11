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

import React, { useContext, useEffect, useState } from 'react';
import { Divider, Form, Result } from 'antd';
import { StationStoreContext } from 'domain/stationOverview';
import { ReactComponent as ConnectorIcon } from 'assets/images/connectorIcon.svg';
import InputNumberComponent from 'components/InputNumber';
import TitleComponent from 'components/titleComponent';
import SelectComponent from 'components/select';
import { Select } from 'antd';
import Input from 'components/Input';
import Modal from 'components/modal';
import Spinner from 'components/spinner';
import { ApiEndpoints } from 'const/apiEndpoints';
import { httpRequest } from 'services/http';
import CloudModal from 'components/cloudModal';
import { isCloud } from 'services/valueConvertor';
import { sendTrace } from 'services/genericServices';
import { connectorTypesSource } from 'connectors';
import { connectorTypesSink } from 'connectors';

const ConnectorModal = ({ open, clickOutside, newConnecor, source }) => {
    const [stationState, stationDispatch] = useContext(StationStoreContext);
    const [connectorForm] = Form.useForm();
    const [step, setStep] = useState(1);
    const [loading, setLoading] = useState(false);
    const [formFields, setFormFields] = useState({
        connector_type: source ? 'Source' : 'Sink'
    });
    const [connectorInputFields, setConnectorInputFields] = useState([]);
    const [resError, setError] = useState(null);
    const [cloudModalopen, setCloudModalOpen] = useState(false);

    useEffect(() => {
        if (open) {
            setStep(1);
            setFormFields({
                connector_type: source ? 'Source' : 'Sink'
            });
            setLoading(false);
            setError(null);
        }
    }, [open]);

    useEffect(() => {
        let connectorType =
            formFields?.connector_type === 'Source'
                ? connectorTypesSource.find((connector) => connector.name === formFields?.type)
                : connectorTypesSink.find((connector) => connector.name === formFields?.type);
        formFields?.type && setConnectorInputFields(connectorType?.inputs[formFields?.connector_type]);
    }, [formFields?.type, formFields?.connector_type]);

    const updateFormState = (key, value) => {
        setFormFields({ ...formFields, [key]: value });
    };
    const updateFormSettingsState = (key, value) => {
        let settings = formFields?.settings || {};
        settings[key] = value;
        setFormFields({ ...formFields, settings });
    };

    const updateMultiFormState = (key, value, index) => {
        let settings = formFields?.settings || {};
        settings[key] = value;
        setFormFields({ ...formFields, settings });
    };

    const connectorModalTitle =
        step === 1 ? (
            <>
                <div className="connector-modal-title">
                    <div className="modal-title">
                        Add a new connector <span className="coming-soon-select">Alpha</span>
                    </div>
                </div>
                <label>Choose a ready-to-use sink for the ingested messages</label>
            </>
        ) : loading ? (
            <>
                <p>Validating connectivity</p>
                <label>Choose a ready-to-use sink for the ingested messages</label>
            </>
        ) : resError ? (
            <>
                <p>Error while creating connector</p>
                <label>Choose a ready-to-use sink for the ingested messages</label>
            </>
        ) : (
            <>
                <p>Created successfully</p>
                <label>Choose a ready-to-use sink for the ingested messages</label>
            </>
        );
    const onFinish = async () => {
        try {
            if (step === 1) {
                try {
                    await connectorForm.validateFields();
                    sendTrace('createConnector', {
                        name: formFields?.name,
                        type: formFields?.type?.toLocaleLowerCase(),
                        connector_type: formFields?.connector_type?.toLocaleLowerCase()
                    });
                    isCloud() ? createConnector() : setCloudModalOpen(true);
                } catch (err) {
                    return;
                }
            } else {
                resError ? setStep(1) : clickOutside();
                setError(null);
            }
        } catch (err) {
            return;
        }
    };

    const createConnector = async () => {
        setLoading(true);
        setError(null);
        setStep(2);
        try {
            const modifiedSettings = { ...formFields?.settings };
            for (const key in modifiedSettings) {
                if (Array.isArray(modifiedSettings[key])) {
                    modifiedSettings[key] = modifiedSettings[key].join(',');
                }
            }
            let data = await httpRequest('POST', ApiEndpoints.CREATE_CONNECTOR, {
                name: formFields?.name,
                station_id: stationState?.stationMetaData?.id,
                type: formFields?.type?.toLocaleLowerCase(),
                connector_type: formFields?.connector_type?.toLocaleLowerCase(),
                settings: modifiedSettings,
                partitions: [stationState?.stationPartition],
                instances: formFields?.instances
            });
            newConnecor(data?.connector, formFields?.connector_type?.toLocaleLowerCase());
        } catch (error) {
            setError(JSON.stringify(error?.data?.message || error?.data));
        } finally {
            setLoading(false);
        }
    };

    const generateFormItem = (input, index, depth, inputName) => {
        return (
            <>
                <Form.Item
                    name={input?.name}
                    validateTrigger="onChange"
                    rules={[
                        {
                            required: input?.required,
                            message: `required`
                        }
                    ]}
                >
                    <TitleComponent headerTitle={input?.display} typeTitle="sub-header" headerDescription={input?.description} required={input?.required} />
                    {input?.type === 'string' && (
                        <Input
                            value={formFields[input?.name]}
                            placeholder={input?.placeholder}
                            type="text"
                            radiusType="semi-round"
                            backgroundColorType="none"
                            boxShadowsType="none"
                            colorType="black"
                            borderColorType="gray"
                            width="100%"
                            height="40px"
                            fontSize="14px"
                            onChange={(e) => {
                                input?.name === 'name' ? updateFormState(input?.name, e.target.value) : updateFormSettingsState(input?.name, e.target.value);
                                connectorForm.setFieldValue(input?.name, e.target.value);
                            }}
                        />
                    )}
                    {input?.type === 'number' && (
                        <InputNumberComponent
                            radiusType="semi-round"
                            backgroundColorType="none"
                            boxShadowsType="none"
                            colorType="black"
                            borderColorType="gray"
                            fontSize="14px"
                            style={{ width: '100%', height: '40px', display: 'flex', alignItems: 'center' }}
                            min={input?.min || 0}
                            max={input?.max || null}
                            onChange={(e) => {
                                input?.name === 'instances' ? updateFormState(input?.name, e) : updateFormSettingsState(input?.name, e);
                                connectorForm.setFieldValue(input?.name, e);
                            }}
                            value={formFields[input?.name]}
                            placeholder={input?.placeholder}
                        />
                    )}
                    {input?.type === 'select' && (
                        <SelectComponent
                            colorType="black"
                            backgroundColorType="none"
                            fontFamily="Inter"
                            borderColorType="gray"
                            radiusType="semi-round"
                            height="40px"
                            popupClassName="select-options"
                            options={input?.options}
                            value={formFields[input?.name]}
                            placeholder={input?.placeholder}
                            onChange={(e) => {
                                updateFormSettingsState(input?.name, e);
                                connectorForm.setFieldValue(input?.name, e);
                            }}
                            disabled={false}
                        />
                    )}
                    {input?.type === 'multi' && (
                        <Select
                            mode="tags"
                            placeholder={input?.placeholder}
                            value={formFields?.settings ? formFields?.settings[input?.name] : []}
                            onChange={(values) => {
                                updateMultiFormState(input?.name, values, index);
                                connectorForm.setFieldValue(input?.name, values);
                            }}
                            style={{ width: '100%' }}
                            onInputValueChange={(value) => {
                                updateMultiFormState(input?.name, value, index);
                            }}
                            popupClassName="select-options"
                        >
                            {formFields?.settings && formFields?.settings[input?.name]?.map((item) => <Select.Option key={item}>{item}</Select.Option>)}
                        </Select>
                    )}
                </Form.Item>

                {depth === 0 &&
                    input?.children &&
                    formFields?.settings &&
                    formFields?.settings[input?.name] &&
                    connectorInputFields[index][formFields?.settings[input?.name]]?.map((child, index) => generateFormItem(child, index, depth + 1, input?.name))}

                {depth === 1 &&
                    input?.children &&
                    formFields?.settings[input?.name] &&
                    input[formFields?.settings[input?.name]]?.map((child, index) => generateFormItem(child, index, depth + 1, input?.name))}
            </>
        );
    };

    return (
        <Modal
            header={
                <div className="modal-header connector-modal-header">
                    <div className="header-img-container">
                        <ConnectorIcon className="headerImage" alt="stationImg" />
                    </div>
                    {connectorModalTitle}
                </div>
            }
            className={'modal-wrapper produce-modal'}
            width="550px"
            clickOutside={clickOutside}
            open={open}
            displayButtons={true}
            rBtnText={step === 1 ? 'Next' : resError ? 'Back' : 'Done'}
            lBtnText={'Close'}
            rBtnClick={onFinish}
            lBtnClick={clickOutside}
            isLoading={false}
            keyListener={false}
        >
            <div className="connector-modal-wrapper">
                {step === 1 && (
                    <Form name="form" form={connectorForm} autoComplete="on" className={'connector-form'}>
                        <Form.Item name="connector_type" validateTrigger="onChange">
                            <TitleComponent headerTitle="Direction" typeTitle="sub-header" headerDescription="Choose the direction of the connector" />
                            <SelectComponent
                                colorType="black"
                                backgroundColorType="none"
                                fontFamily="Inter"
                                borderColorType="gray"
                                radiusType="semi-round"
                                height="40px"
                                popupClassName="select-options"
                                options={['Source', 'Sink']}
                                value={formFields?.connector_type}
                                onChange={(e) => updateFormState('connector_type', e)}
                                onBlur={(e) => updateFormState('connector_type', e)}
                                disabled={false}
                            />
                        </Form.Item>
                        <Form.Item name="type" validateTrigger="onChange">
                            <TitleComponent headerTitle="Type" typeTitle="sub-header" headerDescription="Choose the type of the connector" />
                            <SelectComponent
                                colorType="black"
                                backgroundColorType="none"
                                fontFamily="Inter"
                                borderColorType="gray"
                                radiusType="semi-round"
                                height="40px"
                                popupClassName="select-options"
                                options={formFields?.connector_type === 'Source' ? connectorTypesSource : connectorTypesSink}
                                value={formFields?.type}
                                onChange={(e) => {
                                    updateFormState('type', e);
                                    connectorForm.setFieldValue('connectorType', e);
                                }}
                                disabled={false}
                            />
                        </Form.Item>
                        {formFields?.type && (
                            <>
                                <Divider />
                                <div className="connector-inputs">
                                    {connectorInputFields?.map((input, index) => {
                                        return generateFormItem(input, index, 0);
                                    })}
                                </div>
                            </>
                        )}
                    </Form>
                )}
                {step === 2 && (!resError || (resError && Object.keys(resError)?.length === 0)) && (
                    <div className="validation">
                        {loading ? (
                            <>
                                <Spinner fontSize={60} />
                                <p>Waiting for messages to arrive</p>
                            </>
                        ) : (
                            <Result status="success" subTitle="The connector has been successfully created and is now ready to use" />
                        )}
                    </div>
                )}

                {step === 2 && resError && Object.keys(resError)?.length > 0 && !loading && <div className="result">{resError}</div>}
            </div>
            <CloudModal type="cloud" open={cloudModalopen} handleClose={() => setCloudModalOpen(false)} />
        </Modal>
    );
};

export default ConnectorModal;
