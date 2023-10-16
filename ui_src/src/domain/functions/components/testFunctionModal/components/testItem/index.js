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

import { useState, useEffect } from 'react';
import CheckboxComponent from '../../../../../../components/checkBox';
import { FiEdit } from 'react-icons/fi';
import { ReactComponent as BinIcon } from '../../../../../../assets/images/binIcon.svg';
import { ReactComponent as TestEventModalIcon } from '../../../../../../assets/images/testEventModalcon.svg';

import Modal from '../../../../../../components/modal';
import EditTestEventModal from '../editTestEventModal';
const TestItem = ({ data, handleCheckedClick, isCheck, handleDelete, handleEdit }) => {
    const [isEditModalOpen, setIsEditModalOpen] = useState(false);
    useEffect(() => {
        setIsEditModalOpen(false);
    }, [data]);

    return (
        <>
            <div
                className={`testitem-container ${isCheck ? 'selected' : ''}`}
                tabIndex={0}
                onClick={() => {
                    setIsEditModalOpen(true);
                }}
            >
                <div className="firstRow">
                    <div className="leftGroup">
                        <CheckboxComponent checked={isCheck} id={data.name} onChange={handleCheckedClick} />

                        <p className="title">{data.name}</p>
                    </div>
                    <div className="actions">
                        <FiEdit
                            width={10}
                            height={10}
                            color={'#5C5F62'}
                            alt="edit icon"
                            onClick={(e) => {
                                e.stopPropagation();
                                handleEdit(data.name);
                            }}
                            className="icon"
                        />
                        <BinIcon
                            width={14}
                            height={14}
                            alt="bin icon"
                            onClick={(e) => {
                                e.stopPropagation();
                                handleDelete(data.name);
                            }}
                            className="icon"
                        />
                    </div>
                </div>
                <p className="subtitle">{data.description}</p>
            </div>
            <Modal
                width={'75vw'}
                height={'75vh'}
                clickOutside={() => {
                    setIsEditModalOpen(false);
                }}
                displayButtons={false}
                open={isEditModalOpen}
                header={
                    <div className="modal-header">
                        <div className="header-img-container">
                            <TestEventModalIcon alt="testEventModalIcon" />
                        </div>
                        <p>Create your template</p>
                        <label>Lorem Ipsum is simply dummy text of the printing and typesetting industry.</label>
                    </div>
                }
            >
                <EditTestEventModal
                    handleEdit={() => {
                        handleEdit(data.name);
                        setIsEditModalOpen(false);
                    }}
                    event={data}
                />
            </Modal>
        </>
    );
};

export default TestItem;
