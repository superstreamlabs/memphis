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
import Button from '../../../../components/button';
import { FiChevronRight } from 'react-icons/fi';
import SearchInput from '../../../../components/searchInput';
import { ReactComponent as SearchIcon } from '../../../../assets/images/searchIcon.svg';
import { ReactComponent as TestEventModalIcon } from '../../../../assets/images/testEventModalcon.svg';
import TestItem from './components/testItem';
import Modal from '../../../../components/modal';
import NewTestEventModal from './components/newTestEventModal';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import { httpRequest } from '../../../../services/http';
import { ReactComponent as DeleteWrapperIcon } from '../../../../assets/images/deleteWrapperIcon.svg';
import { ReactComponent as EmptyEventsIcon } from '../../../../assets/images/emptyEvents.svg';
import DeleteItemsModal from '../../../../components/deleteItemsModal';

const TestFunctionModal = ({ onCancel }) => {
    const [searchEvent, setSearchEvent] = useState('');
    const [isNewTestEventModalOpen, setIsNewTestEventModalOpen] = useState(false);
    const [testEvents, setTestEvents] = useState([]);
    const [isCheck, setIsCheck] = useState([]);
    const [isCheckAll, setIsCheckAll] = useState(false);
    const [filteredEvents, setFilteredEvents] = useState([]);
    const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
    const [deleteLoader, setDeleteLoader] = useState(false);
    const [deleteEvent, setDeleteEvent] = useState(null);
    const [editEvent, setEditEvent] = useState(null);

    useEffect(() => {
        getAllTestEvents();
    }, []);

    useEffect(() => {
        if (isCheck.length === filteredEvents.length && filteredEvents.length > 0) {
            setIsCheckAll(true);
        }
    }, [testEvents, isCheck]);

    useEffect(() => {
        let results = testEvents;
        if (searchEvent.length > 0) {
            results = results.filter((testEvent) => testEvent.name.toLowerCase().includes(searchEvent.toLowerCase()));
        }
        setFilteredEvents(results);
    }, [searchEvent, testEvents]);

    const getAllTestEvents = async () => {
        const response = await httpRequest('GET', ApiEndpoints.GET_ALL_TEST_EVENTS);
        setTestEvents(response?.test_events);
    };

    const handleCheckedClick = (e) => {
        const { id, checked } = e.target;
        setIsCheck([...isCheck, id]);
        if (!checked) {
            setIsCheck(isCheck.filter((item) => item !== id));
        }
        if (isCheck.length === 1 && !checked) {
            setIsCheckAll(false);
        }
    };

    const handleDeleteSelected = async (testEvent) => {
        setDeleteLoader(true);
        try {
            if (testEvent) {
                const results = await httpRequest('DELETE', ApiEndpoints.DELETE_TEST_EVENT, { test_event_name: testEvent });
                if (results) {
                    setIsDeleteModalOpen(false);
                    getAllTestEvents();
                    setDeleteEvent(null);
                }
            } else {
                const results = await Promise.all(isCheck.map((testEvent) => httpRequest('DELETE', ApiEndpoints.DELETE_TEST_EVENT, { test_event_name: testEvent })));
                if (results.length === isCheck.length) {
                    setIsDeleteModalOpen(false);
                    setIsCheck([]);
                    setIsCheckAll(false);
                    getAllTestEvents();
                }
            }
        } catch (error) {
        } finally {
            setDeleteLoader(false);
        }
    };

    const onCheckedAll = (e) => {
        setIsCheckAll(!isCheckAll);
        setIsCheck(filteredEvents.map((event) => event.name));
        if (isCheckAll) {
            setIsCheck([]);
        }
    };

    const handleDelete = (name) => {
        setDeleteEvent(name);
        setIsDeleteModalOpen(true);
    };

    const handleEdit = (name) => {
        setEditEvent(name);
        setIsNewTestEventModalOpen(true);
    };

    return (
        <div className="testFunction-wrapper">
            <div className="titleIcon">
                <TestEventModalIcon />
            </div>
            <div className="header">
                <div className="title-container">
                    <p className="title">Generate synthethic data</p>
                    <p className="sub-title">In case you prefer to generate a random test event</p>
                </div>
                <Button
                    fontSize={'16px'}
                    fontWeight={'600'}
                    radiusType={'semi-round'}
                    backgroundColorType={'purple'}
                    colorType={'white'}
                    height={'45px'}
                    width={'200px'}
                    placeholder={
                        <div className="button-content">
                            <span>Create New</span> <FiChevronRight style={{ fontSize: '24px' }} />
                        </div>
                    }
                    onClick={() => setIsNewTestEventModalOpen(true)}
                />
            </div>
            <div className="divider">
                <div className="left-line" />
                <span>or</span>
                <div className="right-line" />
            </div>
            <div className="events-wrapper">
                <div className="events-header">
                    <div className="top-row">
                        <p className="title">Select Saved Event</p>
                        <div className="right-side">
                            <Button
                                height="34px"
                                placeholder={`Delete selected (${isCheck?.length})`}
                                colorType="black"
                                radiusType="circle"
                                backgroundColorType="white"
                                fontSize="12px"
                                fontWeight="600"
                                aria-haspopup="true"
                                boxShadowStyle="float"
                                disabled={isCheck?.length === 0}
                                onClick={() => setIsDeleteModalOpen(true)}
                            />
                            <Button
                                width="131px"
                                height="34px"
                                placeholder={isCheckAll ? 'Unselect all' : 'Select all'}
                                colorType="black"
                                radiusType="circle"
                                backgroundColorType="white"
                                fontSize="12px"
                                fontWeight="600"
                                aria-haspopup="true"
                                boxShadowStyle="float"
                                disabled={filteredEvents?.length === 0}
                                onClick={() => onCheckedAll()}
                            />
                        </div>
                    </div>
                    <SearchInput
                        placeholder={'Quick Search'}
                        iconComponent={<SearchIcon />}
                        width={'100%'}
                        height={'44px'}
                        value={searchEvent}
                        colorType={'black'}
                        onChange={(e) => setSearchEvent(e.target.value)}
                        borderColorType={'gray-light'}
                        borderRadiusType={'semi-round'}
                        backgroundColorType={'gray-light'}
                    />
                </div>
                {filteredEvents.length === 0 ? (
                    <div className="noEvent">
                        <EmptyEventsIcon />
                        <p className="noEvent-title">No Saved Events Found</p>
                        <p className="description">Lorem Ipsum is simply dummy text of the printing and typesetting industry. </p>
                    </div>
                ) : (
                    <div className="eventslist-container">
                        {filteredEvents.map((testEvent) => (
                            <TestItem
                                data={testEvent}
                                isCheck={isCheck?.includes(testEvent.name)}
                                handleCheckedClick={handleCheckedClick}
                                handleDelete={handleDelete}
                                handleEdit={handleEdit}
                            />
                        ))}
                    </div>
                )}
            </div>
            <div className="footer">
                <Button
                    placeholder={'Cancel'}
                    backgroundColorType={'white'}
                    border={'gray-light'}
                    colorType={'black'}
                    fontSize={'14px'}
                    fontFamily={'InterSemibold'}
                    radiusType={'circle'}
                    width={'168px'}
                    height={'34px'}
                    onClick={onCancel}
                />
            </div>
            <Modal
                width={'75vw'}
                height={'75vh'}
                clickOutside={() => {
                    setIsNewTestEventModalOpen(false);
                    editEvent && setEditEvent(null);
                }}
                open={isNewTestEventModalOpen}
                displayButtons={false}
                header={
                    <div className="modal-header">
                        <div className="header-img-container">
                            <TestEventModalIcon alt="testEventModalIcon" />
                        </div>
                        <p>Create a test event</p>
                        <label>Test events are made to enable function tests before activation.</label>
                    </div>
                }
            >
                <NewTestEventModal
                    onCancel={() => {
                        setIsNewTestEventModalOpen(false);
                        editEvent && setEditEvent(null);
                    }}
                    updateTestEvents={() => getAllTestEvents()}
                    editEvent={editEvent}
                />
            </Modal>
            <Modal
                header={<DeleteWrapperIcon alt="deleteWrapperIcon" />}
                width="520px"
                height="240px"
                displayButtons={false}
                clickOutside={() => setIsDeleteModalOpen(false)}
                open={isDeleteModalOpen}
            >
                <DeleteItemsModal
                    title="Are you sure you want to delete the selected events?"
                    desc="Deleting these events means they will be permanently deleted."
                    buttontxt="I understand, delete the selected events"
                    handleDeleteSelected={() => handleDeleteSelected(deleteEvent)}
                    loader={deleteLoader}
                />
            </Modal>
        </div>
    );
};

export default TestFunctionModal;
